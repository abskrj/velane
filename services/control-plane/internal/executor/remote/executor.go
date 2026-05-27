package remote

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/runeforge/control-plane/internal/executor"
)

// runRequest is the JSON body sent to executor containers.
type runRequest struct {
	Code        string `json:"code"`
	Input       string `json:"input"`
	TimeoutMs   int    `json:"timeout_ms"`
	MaxMemoryMB int    `json:"max_memory_mb"`
}

// runResponse is the JSON body returned by executor containers.
type runResponse struct {
	Output       string `json:"output"`
	Stderr       string `json:"stderr"`
	DurationMs   int    `json:"duration_ms"`
	PeakMemoryMB int    `json:"peak_memory_mb"`
	ExitCode     int    `json:"exit_code"`
	Error        string `json:"error"`
}

// RemoteExecutor routes execution requests to the appropriate language
// executor container over HTTP.
type RemoteExecutor struct {
	bunEndpoint    string // e.g. "http://bun-executor:8080"
	pythonEndpoint string // e.g. "http://python-executor:8080"
	httpClient     *http.Client
}

// New creates a RemoteExecutor with the supplied endpoint URLs.
func New(bunEndpoint, pythonEndpoint string) *RemoteExecutor {
	return &RemoteExecutor{
		bunEndpoint:    bunEndpoint,
		pythonEndpoint: pythonEndpoint,
		httpClient: &http.Client{
			// The outer context carries the real deadline; this is a backstop.
			Timeout: 5 * time.Minute,
		},
	}
}

// Run sends the spec to the appropriate executor container and returns the
// result. HTTP and JSON errors are surfaced in RunResult.Error.
func (e *RemoteExecutor) Run(ctx context.Context, spec executor.RunSpec) executor.RunResult {
	endpoint, err := e.endpointFor(spec.Language)
	if err != nil {
		return executor.RunResult{Error: err.Error(), ExitCode: -1}
	}

	reqBody, err := json.Marshal(runRequest{
		Code:        spec.Code,
		Input:       spec.Input,
		TimeoutMs:   spec.TimeoutMs,
		MaxMemoryMB: spec.MaxMemoryMB,
	})
	if err != nil {
		return executor.RunResult{Error: fmt.Sprintf("marshal request: %v", err), ExitCode: -1}
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint+"/run", bytes.NewReader(reqBody))
	if err != nil {
		return executor.RunResult{Error: fmt.Sprintf("build request: %v", err), ExitCode: -1}
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := e.httpClient.Do(httpReq)
	if err != nil {
		return executor.RunResult{Error: fmt.Sprintf("executor http: %v", err), ExitCode: -1}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return executor.RunResult{Error: fmt.Sprintf("read body: %v", err), ExitCode: -1}
	}

	if resp.StatusCode >= 500 {
		return executor.RunResult{
			Error:    fmt.Sprintf("executor returned %d: %s", resp.StatusCode, string(body)),
			ExitCode: -1,
		}
	}

	var runResp runResponse
	if err := json.Unmarshal(body, &runResp); err != nil {
		return executor.RunResult{Error: fmt.Sprintf("unmarshal response: %v", err), ExitCode: -1}
	}

	return executor.RunResult{
		Output:       runResp.Output,
		Stderr:       runResp.Stderr,
		DurationMs:   runResp.DurationMs,
		PeakMemoryMB: runResp.PeakMemoryMB,
		ExitCode:     runResp.ExitCode,
		Error:        runResp.Error,
	}
}

func (e *RemoteExecutor) endpointFor(language string) (string, error) {
	switch language {
	case "bun":
		return e.bunEndpoint, nil
	case "python":
		return e.pythonEndpoint, nil
	default:
		return "", fmt.Errorf("unsupported language: %s", language)
	}
}

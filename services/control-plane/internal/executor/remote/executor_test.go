package remote_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/runeforge/control-plane/internal/executor"
	"github.com/runeforge/control-plane/internal/executor/remote"
)

// mockExecutorServer starts an httptest server that responds with the given
// runResponse JSON. Close the returned server after the test.
func mockExecutorServer(t *testing.T, status int, resp any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/run" || r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(resp)
	}))
}

func TestRemoteExecutor_Run_BunSuccess(t *testing.T) {
	srv := mockExecutorServer(t, http.StatusOK, map[string]any{
		"output":         `{"result":"hello"}`,
		"stderr":         "",
		"duration_ms":    42,
		"peak_memory_mb": 12,
		"exit_code":      0,
		"error":          "",
	})
	defer srv.Close()

	exec := remote.New(srv.URL, "http://unused")
	result := exec.Run(context.Background(), executor.RunSpec{
		Language:    "bun",
		Code:        `export default async () => ({ result: "hello" })`,
		Input:       `{}`,
		TimeoutMs:   5000,
		MaxMemoryMB: 128,
	})

	if result.Error != "" {
		t.Fatalf("unexpected error: %v", result.Error)
	}
	if result.Output != `{"result":"hello"}` {
		t.Errorf("output = %q; want %q", result.Output, `{"result":"hello"}`)
	}
	if result.DurationMs != 42 {
		t.Errorf("duration_ms = %d; want 42", result.DurationMs)
	}
	if result.ExitCode != 0 {
		t.Errorf("exit_code = %d; want 0", result.ExitCode)
	}
}

func TestRemoteExecutor_Run_PythonSuccess(t *testing.T) {
	srv := mockExecutorServer(t, http.StatusOK, map[string]any{
		"output":         `{"tokens":10}`,
		"stderr":         "",
		"duration_ms":    80,
		"peak_memory_mb": 30,
		"exit_code":      0,
		"error":          "",
	})
	defer srv.Close()

	exec := remote.New("http://unused", srv.URL)
	result := exec.Run(context.Background(), executor.RunSpec{
		Language:    "python",
		Code:        "async def handler(req): return {'tokens': 10}",
		Input:       `{}`,
		TimeoutMs:   5000,
		MaxMemoryMB: 128,
	})

	if result.Error != "" {
		t.Fatalf("unexpected error: %v", result.Error)
	}
	if result.Output != `{"tokens":10}` {
		t.Errorf("output = %q; want %q", result.Output, `{"tokens":10}`)
	}
}

func TestRemoteExecutor_Run_TimeoutPropagated(t *testing.T) {
	// Server that sleeps longer than the context allows.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"output":"","stderr":"","duration_ms":0,"peak_memory_mb":0,"exit_code":0,"error":""}`))
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	exec := remote.New(srv.URL, "http://unused")
	result := exec.Run(ctx, executor.RunSpec{
		Language:    "bun",
		Code:        "",
		Input:       `{}`,
		TimeoutMs:   50,
		MaxMemoryMB: 128,
	})

	if result.Error == "" {
		t.Error("expected an error due to context timeout, got none")
	}
}

func TestRemoteExecutor_Run_ServerError(t *testing.T) {
	srv := mockExecutorServer(t, http.StatusInternalServerError, map[string]any{
		"error": "executor crashed",
	})
	defer srv.Close()

	exec := remote.New(srv.URL, "http://unused")
	result := exec.Run(context.Background(), executor.RunSpec{
		Language: "bun",
		Input:    `{}`,
	})

	if result.Error == "" {
		t.Error("expected error for 500 response, got none")
	}
	if result.ExitCode != -1 {
		t.Errorf("exit_code = %d; want -1 for server errors", result.ExitCode)
	}
}

func TestRemoteExecutor_Run_UnsupportedLanguage(t *testing.T) {
	exec := remote.New("http://bun", "http://python")
	result := exec.Run(context.Background(), executor.RunSpec{
		Language: "ruby",
		Input:    `{}`,
	})

	if result.Error == "" {
		t.Error("expected error for unsupported language, got none")
	}
	if result.ExitCode != -1 {
		t.Errorf("exit_code = %d; want -1", result.ExitCode)
	}
}

func TestRemoteExecutor_Run_ExecutorReturnsTimeout(t *testing.T) {
	srv := mockExecutorServer(t, http.StatusOK, map[string]any{
		"output":         "",
		"stderr":         "Killed: 9",
		"duration_ms":    5000,
		"peak_memory_mb": 0,
		"exit_code":      1,
		"error":          "timeout",
	})
	defer srv.Close()

	exec := remote.New(srv.URL, "http://unused")
	result := exec.Run(context.Background(), executor.RunSpec{
		Language:  "bun",
		Input:     `{}`,
		TimeoutMs: 5000,
	})

	if result.Error != "timeout" {
		t.Errorf("error = %q; want %q", result.Error, "timeout")
	}
}

func TestRemoteExecutor_Run_InvalidJSONResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`not json`))
	}))
	defer srv.Close()

	exec := remote.New(srv.URL, "http://unused")
	result := exec.Run(context.Background(), executor.RunSpec{
		Language: "bun",
		Input:    `{}`,
	})

	if result.Error == "" {
		t.Error("expected error for invalid JSON response, got none")
	}
}

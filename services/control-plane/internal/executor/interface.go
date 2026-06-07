package executor

import "context"

// EgressPolicy defines the network egress restrictions to enforce during
// snippet execution.
type EgressPolicy struct {
	BlockedCIDRs   []string `json:"blocked_cidrs"`
	BlockedDomains []string `json:"blocked_domains"`
}

// RunSpec describes a single snippet execution request sent to an executor.
type RunSpec struct {
	Language      string            // "bun" | "python"
	Code          string            // raw source code of the snippet
	Input         string            // raw JSON input payload
	TimeoutMs     int               // execution deadline in milliseconds
	MaxMemoryMB   int               // soft memory ceiling in MiB
	SecretEnvVars map[string]string // injected as env vars into the snippet process
	Libraries     map[string]string // importPath → source; written to temp workspace before execution
	EgressPolicy  *EgressPolicy     // nil = no policy enforcement
}

// RunResult captures the outcome of a snippet execution.
type RunResult struct {
	Output       string // raw JSON written to stdout by the snippet
	Stderr       string // anything written to stderr
	DurationMs   int    // wall-clock execution time in milliseconds
	PeakMemoryMB int    // peak RSS in MiB (best-effort)
	ExitCode     int    // OS exit code of the subprocess
	Error        string // "timeout" | "oom" | "" for other errors, or error message
}

// StreamChunk is one typed event emitted by a streaming snippet.
//
// The runtime emits a stream of these as NDJSON / SSE. The Type field
// discriminates the event:
//
//   - "log":    debug output (print / console.log / stderr). Stream is
//     "stdout" or "stderr"; Text holds the line. Dev-gated by the worker.
//   - "chunk":  intentional generator output. Data holds the payload.
//   - "result": the handler's return value. Output holds the raw JSON.
//   - "error":  a terminal failure. Message holds the reason.
//   - "done":   the final sentinel; always last.
//
// Legacy producers that emit only {"data","done","error"} (no "type") are
// still understood: such events are treated as chunk/result by consumers.
type StreamChunk struct {
	Type       string `json:"type,omitempty"`   // log|chunk|result|error|done
	Stream     string `json:"stream,omitempty"` // stdout|stderr (for logs)
	Text       string `json:"text,omitempty"`   // log line text
	Data       string `json:"data,omitempty"`   // chunk payload (partial output)
	Output     string `json:"output,omitempty"` // final result raw JSON
	Message    string `json:"message,omitempty"`
	ExitCode   int    `json:"exit_code,omitempty"`
	DurationMs int    `json:"duration_ms,omitempty"`
	Error      string `json:"error,omitempty"` // legacy/terminal error
	Done       bool   `json:"done,omitempty"`  // true on the final event
}

// Executor is the interface that all execution backends must satisfy.
type Executor interface {
	// Run executes a snippet synchronously and returns the full result.
	Run(ctx context.Context, spec RunSpec) RunResult

	// RunStream calls the executor's streaming endpoint and returns a channel
	// of StreamChunks. The channel is closed when the stream ends or the
	// context is cancelled. The caller must drain the channel.
	RunStream(ctx context.Context, spec RunSpec) (<-chan StreamChunk, error)
}

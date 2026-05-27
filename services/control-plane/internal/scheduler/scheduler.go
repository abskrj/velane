package scheduler

import (
	"context"
	"fmt"

	"github.com/runeforge/control-plane/internal/executor"
	"github.com/runeforge/control-plane/internal/models"
)

// Store is the subset of *postgres.Store that the Scheduler depends on.
// Keeping this narrow makes the scheduler straightforward to test with a mock.
type Store interface {
	GetSnippetBySlug(ctx context.Context, tenantID, slug string) (*models.Snippet, error)
	GetSnippetEnvironment(ctx context.Context, snippetID, env string) (*models.SnippetEnvironment, error)
	GetVersion(ctx context.Context, id string) (*models.SnippetVersion, error)
	CreateInvocation(ctx context.Context, snippetID, versionID, env, tenantID, input string) (*models.Invocation, error)
	UpdateInvocationResult(ctx context.Context, id string, status models.InvocationStatus, output, errMsg, stderr string, durationMs, peakMemoryMB int) error
	GetInvocation(ctx context.Context, id string) (*models.Invocation, error)
}

// InvokeRequest carries the parameters for a synchronous snippet invocation.
type InvokeRequest struct {
	TenantID    string
	SnippetSlug string
	Env         string // "dev" | "prod"
	Input       string // raw JSON
}

// Scheduler resolves, executes, and records snippet invocations.
// In Phase 1 this is a direct pass-through — no warm pool, no queue.
type Scheduler struct {
	store    Store
	executor executor.Executor
}

// New creates a Scheduler wired to the given store and executor.
func New(store Store, exec executor.Executor) *Scheduler {
	return &Scheduler{store: store, executor: exec}
}

// Invoke executes a snippet synchronously:
//  1. Resolve the snippet and active version for the given environment.
//  2. Create an invocation record (status=running).
//  3. Call the executor.
//  4. Persist the result and return the completed invocation.
func (s *Scheduler) Invoke(ctx context.Context, req InvokeRequest) (*models.Invocation, error) {
	// 1. Resolve snippet.
	snippet, err := s.store.GetSnippetBySlug(ctx, req.TenantID, req.SnippetSlug)
	if err != nil {
		return nil, fmt.Errorf("snippet not found: %w", err)
	}

	// Resolve the active version for the requested environment.
	env, err := s.store.GetSnippetEnvironment(ctx, snippet.ID, req.Env)
	if err != nil {
		return nil, fmt.Errorf("get environment: %w", err)
	}
	if env.ActiveVersionID == nil {
		return nil, fmt.Errorf("no published version in environment %q for snippet %q", req.Env, req.SnippetSlug)
	}

	version, err := s.store.GetVersion(ctx, *env.ActiveVersionID)
	if err != nil {
		return nil, fmt.Errorf("get version: %w", err)
	}

	// 2. Record the invocation.
	input := req.Input
	if input == "" {
		input = "{}"
	}

	invocation, err := s.store.CreateInvocation(ctx, snippet.ID, version.ID, req.Env, req.TenantID, input)
	if err != nil {
		return nil, fmt.Errorf("create invocation: %w", err)
	}

	// 3. Execute.
	result := s.executor.Run(ctx, executor.RunSpec{
		Language:    string(snippet.Language),
		Code:        version.Code,
		Input:       input,
		TimeoutMs:   version.TimeoutMs,
		MaxMemoryMB: version.MaxMemoryMB,
	})

	// 4. Persist result.
	status := models.InvocationCompleted
	if result.Error == "timeout" {
		status = models.InvocationTimeout
	} else if result.Error == "oom" {
		status = models.InvocationOOMKilled
	} else if result.ExitCode != 0 || result.Error != "" {
		status = models.InvocationFailed
	}

	updateErr := s.store.UpdateInvocationResult(ctx,
		invocation.ID,
		status,
		result.Output,
		result.Error,
		result.Stderr,
		result.DurationMs,
		result.PeakMemoryMB,
	)
	if updateErr != nil {
		// Log-worthy but don't fail the caller — they still get the result.
		_ = updateErr
	}

	// Return the updated invocation.
	final, err := s.store.GetInvocation(ctx, invocation.ID)
	if err != nil {
		// Fall back to assembling the struct manually.
		invocation.Status = status
		invocation.Output = result.Output
		invocation.Error = result.Error
		invocation.Stderr = result.Stderr
		invocation.DurationMs = result.DurationMs
		invocation.PeakMemoryMB = result.PeakMemoryMB
		return invocation, nil
	}

	return final, nil
}

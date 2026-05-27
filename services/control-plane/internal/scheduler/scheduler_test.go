package scheduler_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/runeforge/control-plane/internal/executor"
	"github.com/runeforge/control-plane/internal/models"
	"github.com/runeforge/control-plane/internal/scheduler"
)

// --- Mocks ---

type mockStore struct {
	getSnippetBySlug       func(ctx context.Context, tenantID, slug string) (*models.Snippet, error)
	getSnippetEnvironment  func(ctx context.Context, snippetID, env string) (*models.SnippetEnvironment, error)
	getVersion             func(ctx context.Context, id string) (*models.SnippetVersion, error)
	createInvocation       func(ctx context.Context, snippetID, versionID, env, tenantID, input string) (*models.Invocation, error)
	updateInvocationResult func(ctx context.Context, id string, status models.InvocationStatus, output, errMsg, stderr string, durationMs, peakMemoryMB int) error
	getInvocation          func(ctx context.Context, id string) (*models.Invocation, error)
}

func (m *mockStore) GetSnippetBySlug(ctx context.Context, tenantID, slug string) (*models.Snippet, error) {
	return m.getSnippetBySlug(ctx, tenantID, slug)
}
func (m *mockStore) GetSnippetEnvironment(ctx context.Context, snippetID, env string) (*models.SnippetEnvironment, error) {
	return m.getSnippetEnvironment(ctx, snippetID, env)
}
func (m *mockStore) GetVersion(ctx context.Context, id string) (*models.SnippetVersion, error) {
	return m.getVersion(ctx, id)
}
func (m *mockStore) CreateInvocation(ctx context.Context, snippetID, versionID, env, tenantID, input string) (*models.Invocation, error) {
	return m.createInvocation(ctx, snippetID, versionID, env, tenantID, input)
}
func (m *mockStore) UpdateInvocationResult(ctx context.Context, id string, status models.InvocationStatus, output, errMsg, stderr string, durationMs, peakMemoryMB int) error {
	return m.updateInvocationResult(ctx, id, status, output, errMsg, stderr, durationMs, peakMemoryMB)
}
func (m *mockStore) GetInvocation(ctx context.Context, id string) (*models.Invocation, error) {
	return m.getInvocation(ctx, id)
}

type mockExecutor struct {
	run func(ctx context.Context, spec executor.RunSpec) executor.RunResult
}

func (m *mockExecutor) Run(ctx context.Context, spec executor.RunSpec) executor.RunResult {
	return m.run(ctx, spec)
}

// --- Fixtures ---

func fixSnippet() *models.Snippet {
	return &models.Snippet{
		ID:       "snip-1",
		TenantID: "tenant-1",
		Slug:     "hello",
		Language: models.LanguageBun,
	}
}

func fixVersion(id string) *models.SnippetVersion {
	return &models.SnippetVersion{
		ID:          id,
		SnippetID:   "snip-1",
		Code:        `export async function handler() { return {ok:true} }`,
		TimeoutMs:   5000,
		MaxMemoryMB: 128,
		Status:      models.StatusPublished,
	}
}

func fixInvocation(id, versionID string) *models.Invocation {
	now := time.Now()
	return &models.Invocation{
		ID:          id,
		SnippetID:   "snip-1",
		VersionID:   versionID,
		Environment: "prod",
		TenantID:    "tenant-1",
		Status:      models.InvocationRunning,
		CreatedAt:   now,
	}
}

// --- Tests ---

func TestScheduler_Invoke_Success(t *testing.T) {
	versionID := "ver-1"
	invocationID := "inv-1"
	activeVersion := versionID

	store := &mockStore{
		getSnippetBySlug: func(ctx context.Context, tenantID, slug string) (*models.Snippet, error) {
			return fixSnippet(), nil
		},
		getSnippetEnvironment: func(ctx context.Context, snippetID, env string) (*models.SnippetEnvironment, error) {
			return &models.SnippetEnvironment{
				SnippetID:       snippetID,
				Env:             env,
				ActiveVersionID: &activeVersion,
			}, nil
		},
		getVersion: func(ctx context.Context, id string) (*models.SnippetVersion, error) {
			return fixVersion(id), nil
		},
		createInvocation: func(ctx context.Context, snippetID, versionID, env, tenantID, input string) (*models.Invocation, error) {
			return fixInvocation(invocationID, versionID), nil
		},
		updateInvocationResult: func(ctx context.Context, id string, status models.InvocationStatus, output, errMsg, stderr string, durationMs, peakMemoryMB int) error {
			return nil
		},
		getInvocation: func(ctx context.Context, id string) (*models.Invocation, error) {
			inv := fixInvocation(invocationID, versionID)
			inv.Status = models.InvocationCompleted
			inv.Output = `{"ok":true}`
			inv.DurationMs = 42
			return inv, nil
		},
	}

	exec := &mockExecutor{
		run: func(ctx context.Context, spec executor.RunSpec) executor.RunResult {
			return executor.RunResult{
				Output:     `{"ok":true}`,
				DurationMs: 42,
				ExitCode:   0,
			}
		},
	}

	sched := scheduler.New(store, exec)
	inv, err := sched.Invoke(context.Background(), scheduler.InvokeRequest{
		TenantID:    "tenant-1",
		SnippetSlug: "hello",
		Env:         "prod",
		Input:       `{}`,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv.Status != models.InvocationCompleted {
		t.Errorf("status = %q; want %q", inv.Status, models.InvocationCompleted)
	}
	if inv.Output != `{"ok":true}` {
		t.Errorf("output = %q; want %q", inv.Output, `{"ok":true}`)
	}
}

func TestScheduler_Invoke_SnippetNotFound(t *testing.T) {
	store := &mockStore{
		getSnippetBySlug: func(ctx context.Context, tenantID, slug string) (*models.Snippet, error) {
			return nil, errors.New("not found")
		},
	}

	sched := scheduler.New(store, &mockExecutor{})
	_, err := sched.Invoke(context.Background(), scheduler.InvokeRequest{
		TenantID:    "tenant-1",
		SnippetSlug: "missing",
		Env:         "prod",
	})

	if err == nil {
		t.Fatal("expected error for missing snippet, got nil")
	}
}

func TestScheduler_Invoke_NoPublishedVersion(t *testing.T) {
	store := &mockStore{
		getSnippetBySlug: func(ctx context.Context, tenantID, slug string) (*models.Snippet, error) {
			return fixSnippet(), nil
		},
		getSnippetEnvironment: func(ctx context.Context, snippetID, env string) (*models.SnippetEnvironment, error) {
			return &models.SnippetEnvironment{
				SnippetID:       snippetID,
				Env:             env,
				ActiveVersionID: nil, // nothing published
			}, nil
		},
	}

	sched := scheduler.New(store, &mockExecutor{})
	_, err := sched.Invoke(context.Background(), scheduler.InvokeRequest{
		TenantID:    "tenant-1",
		SnippetSlug: "hello",
		Env:         "prod",
	})

	if err == nil {
		t.Fatal("expected error when no version is published, got nil")
	}
}

func TestScheduler_Invoke_ExecutorTimeout(t *testing.T) {
	versionID := "ver-1"
	invocationID := "inv-1"
	activeVersion := versionID

	var capturedStatus models.InvocationStatus
	store := &mockStore{
		getSnippetBySlug: func(ctx context.Context, tenantID, slug string) (*models.Snippet, error) {
			return fixSnippet(), nil
		},
		getSnippetEnvironment: func(ctx context.Context, snippetID, env string) (*models.SnippetEnvironment, error) {
			return &models.SnippetEnvironment{
				SnippetID:       snippetID,
				Env:             env,
				ActiveVersionID: &activeVersion,
			}, nil
		},
		getVersion: func(ctx context.Context, id string) (*models.SnippetVersion, error) {
			return fixVersion(id), nil
		},
		createInvocation: func(ctx context.Context, snippetID, versionID, env, tenantID, input string) (*models.Invocation, error) {
			return fixInvocation(invocationID, versionID), nil
		},
		updateInvocationResult: func(ctx context.Context, id string, status models.InvocationStatus, output, errMsg, stderr string, durationMs, peakMemoryMB int) error {
			capturedStatus = status
			return nil
		},
		getInvocation: func(ctx context.Context, id string) (*models.Invocation, error) {
			inv := fixInvocation(invocationID, versionID)
			inv.Status = capturedStatus
			return inv, nil
		},
	}

	exec := &mockExecutor{
		run: func(ctx context.Context, spec executor.RunSpec) executor.RunResult {
			return executor.RunResult{
				Error:    "timeout",
				ExitCode: 1,
			}
		},
	}

	sched := scheduler.New(store, exec)
	inv, err := sched.Invoke(context.Background(), scheduler.InvokeRequest{
		TenantID:    "tenant-1",
		SnippetSlug: "hello",
		Env:         "prod",
		Input:       `{}`,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv.Status != models.InvocationTimeout {
		t.Errorf("status = %q; want %q", inv.Status, models.InvocationTimeout)
	}
}

func TestScheduler_Invoke_ExecutorOOM(t *testing.T) {
	versionID := "ver-1"
	invocationID := "inv-1"
	activeVersion := versionID

	var capturedStatus models.InvocationStatus
	store := &mockStore{
		getSnippetBySlug: func(ctx context.Context, tenantID, slug string) (*models.Snippet, error) {
			return fixSnippet(), nil
		},
		getSnippetEnvironment: func(ctx context.Context, snippetID, env string) (*models.SnippetEnvironment, error) {
			return &models.SnippetEnvironment{
				SnippetID: snippetID, Env: env, ActiveVersionID: &activeVersion,
			}, nil
		},
		getVersion: func(ctx context.Context, id string) (*models.SnippetVersion, error) {
			return fixVersion(id), nil
		},
		createInvocation: func(ctx context.Context, snippetID, versionID, env, tenantID, input string) (*models.Invocation, error) {
			return fixInvocation(invocationID, versionID), nil
		},
		updateInvocationResult: func(ctx context.Context, id string, status models.InvocationStatus, output, errMsg, stderr string, durationMs, peakMemoryMB int) error {
			capturedStatus = status
			return nil
		},
		getInvocation: func(ctx context.Context, id string) (*models.Invocation, error) {
			inv := fixInvocation(invocationID, versionID)
			inv.Status = capturedStatus
			return inv, nil
		},
	}

	exec := &mockExecutor{
		run: func(ctx context.Context, spec executor.RunSpec) executor.RunResult {
			return executor.RunResult{Error: "oom", ExitCode: 137}
		},
	}

	sched := scheduler.New(store, exec)
	inv, err := sched.Invoke(context.Background(), scheduler.InvokeRequest{
		TenantID: "tenant-1", SnippetSlug: "hello", Env: "prod", Input: `{}`,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv.Status != models.InvocationOOMKilled {
		t.Errorf("status = %q; want %q", inv.Status, models.InvocationOOMKilled)
	}
}

func TestScheduler_Invoke_DefaultsEmptyInputToEmptyObject(t *testing.T) {
	versionID := "ver-1"
	invocationID := "inv-1"
	activeVersion := versionID

	var capturedInput string
	store := &mockStore{
		getSnippetBySlug: func(ctx context.Context, tenantID, slug string) (*models.Snippet, error) {
			return fixSnippet(), nil
		},
		getSnippetEnvironment: func(ctx context.Context, snippetID, env string) (*models.SnippetEnvironment, error) {
			return &models.SnippetEnvironment{
				SnippetID: snippetID, Env: env, ActiveVersionID: &activeVersion,
			}, nil
		},
		getVersion: func(ctx context.Context, id string) (*models.SnippetVersion, error) {
			return fixVersion(id), nil
		},
		createInvocation: func(ctx context.Context, snippetID, versionID, env, tenantID, input string) (*models.Invocation, error) {
			capturedInput = input
			return fixInvocation(invocationID, versionID), nil
		},
		updateInvocationResult: func(ctx context.Context, id string, status models.InvocationStatus, output, errMsg, stderr string, durationMs, peakMemoryMB int) error {
			return nil
		},
		getInvocation: func(ctx context.Context, id string) (*models.Invocation, error) {
			return fixInvocation(invocationID, versionID), nil
		},
	}

	exec := &mockExecutor{
		run: func(ctx context.Context, spec executor.RunSpec) executor.RunResult {
			return executor.RunResult{Output: `{}`, ExitCode: 0}
		},
	}

	sched := scheduler.New(store, exec)
	_, err := sched.Invoke(context.Background(), scheduler.InvokeRequest{
		TenantID: "tenant-1", SnippetSlug: "hello", Env: "prod",
		Input: "", // empty — should default to "{}"
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedInput != "{}" {
		t.Errorf("capturedInput = %q; want %q", capturedInput, "{}")
	}
}

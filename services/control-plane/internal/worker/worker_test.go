package worker_test

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/abskrj/velane/services/control-plane/internal/executor"
	"github.com/abskrj/velane/services/control-plane/internal/models"
	redisstore "github.com/abskrj/velane/services/control-plane/internal/store/redis"
	"github.com/abskrj/velane/services/control-plane/internal/worker"
	"go.uber.org/zap"
)

// --- Mocks ---

// mockDequeuer implements worker.Dequeuer by returning jobs from a buffered channel.
type mockDequeuer struct {
	jobs chan *redisstore.Job
}

func newMockDequeuer(jobs ...*redisstore.Job) *mockDequeuer {
	ch := make(chan *redisstore.Job, len(jobs)+1)
	for _, j := range jobs {
		ch <- j
	}
	return &mockDequeuer{jobs: ch}
}

func (m *mockDequeuer) Dequeue(ctx context.Context) (*redisstore.Job, error) {
	select {
	case <-ctx.Done():
		return nil, nil
	case j, ok := <-m.jobs:
		if !ok {
			<-ctx.Done()
			return nil, nil
		}
		return j, nil
	}
}

// mockWorkerStore captures UpdateInvocationResult calls.
type mockWorkerStore struct {
	mu      sync.Mutex
	results []capturedResult
}

type capturedResult struct {
	id     string
	status models.InvocationStatus
	output string
	errMsg string
}

func (m *mockWorkerStore) UpdateInvocationResult(
	ctx context.Context, id string, status models.InvocationStatus,
	output, errMsg, stderr string, durationMs, peakMemoryMB, cpuMs int,
) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.results = append(m.results, capturedResult{
		id:     id,
		status: status,
		output: output,
		errMsg: errMsg,
	})
	return nil
}

func (m *mockWorkerStore) latest() *capturedResult {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.results) == 0 {
		return nil
	}
	r := m.results[len(m.results)-1]
	return &r
}

func (m *mockWorkerStore) count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.results)
}

// mockExecutor implements executor.Executor with configurable behaviour.
type mockExecutor struct {
	run       func(ctx context.Context, spec executor.RunSpec) executor.RunResult
	runStream func(ctx context.Context, spec executor.RunSpec) (<-chan executor.StreamChunk, error)
}

func (m *mockExecutor) Run(ctx context.Context, spec executor.RunSpec) executor.RunResult {
	if m.run != nil {
		return m.run(ctx, spec)
	}
	return executor.RunResult{}
}

func (m *mockExecutor) RunStream(ctx context.Context, spec executor.RunSpec) (<-chan executor.StreamChunk, error) {
	if m.runStream != nil {
		return m.runStream(ctx, spec)
	}
	ch := make(chan executor.StreamChunk)
	close(ch)
	return ch, nil
}

// mockPublisher captures events published to invocation event streams.
type mockPublisher struct {
	mu     sync.Mutex
	events map[string][]string
}

func newMockPublisher() *mockPublisher {
	return &mockPublisher{events: make(map[string][]string)}
}

func (m *mockPublisher) PublishEvent(_ context.Context, invocationID string, payload []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events[invocationID] = append(m.events[invocationID], string(payload))
	return nil
}

func (m *mockPublisher) forID(id string) []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]string, len(m.events[id]))
	copy(out, m.events[id])
	return out
}

// runStreamWorker runs a single streaming job to completion.
func runStreamWorker(t *testing.T, deq *mockDequeuer, store *mockWorkerStore, exec executor.Executor, pub worker.EventPublisher) {
	t.Helper()
	log := zap.NewNop()
	w := worker.New(deq, store, exec, log, 1)
	w.SetEventPublisher(pub)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		w.Run(ctx)
	}()

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if store.count() >= 1 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Error("worker did not stop within 2s after cancel")
	}
}

func streamExecutor(events []executor.StreamChunk) *mockExecutor {
	return &mockExecutor{
		runStream: func(_ context.Context, _ executor.RunSpec) (<-chan executor.StreamChunk, error) {
			ch := make(chan executor.StreamChunk, len(events))
			for _, e := range events {
				ch <- e
			}
			close(ch)
			return ch, nil
		},
	}
}

func countEventType(events []string, typ string) int {
	n := 0
	for _, e := range events {
		var v struct {
			Type string `json:"type"`
		}
		_ = json.Unmarshal([]byte(e), &v)
		if v.Type == typ {
			n++
		}
	}
	return n
}

func TestWorker_StreamJob_DevForwardsLogs(t *testing.T) {
	job := makeJob("inv-stream-dev")
	job.Stream = true
	job.Env = "dev"

	exec := streamExecutor([]executor.StreamChunk{
		{Type: "log", Stream: "stdout", Text: "hello"},
		{Type: "result", Output: `{"ok":true}`},
		{Type: "done", Done: true},
	})
	store := &mockWorkerStore{}
	pub := newMockPublisher()

	runStreamWorker(t, newMockDequeuer(job), store, exec, pub)

	res := store.latest()
	if res == nil || res.status != models.InvocationCompleted {
		t.Fatalf("expected completed status, got %+v", res)
	}
	if res.output != `{"ok":true}` {
		t.Errorf("output = %q; want %q", res.output, `{"ok":true}`)
	}

	events := pub.forID(job.InvocationID)
	if countEventType(events, "log") != 1 {
		t.Errorf("expected 1 log event in dev, got %d (events=%v)", countEventType(events, "log"), events)
	}
	if countEventType(events, "done") != 1 {
		t.Errorf("expected exactly 1 terminal done event, got %d", countEventType(events, "done"))
	}
}

func TestWorker_StreamJob_ProdDropsLogs(t *testing.T) {
	job := makeJob("inv-stream-prod")
	job.Stream = true
	job.Env = "prod"

	exec := streamExecutor([]executor.StreamChunk{
		{Type: "log", Stream: "stdout", Text: "secret debug line"},
		{Type: "result", Output: `{"ok":true}`},
		{Type: "done", Done: true},
	})
	store := &mockWorkerStore{}
	pub := newMockPublisher()

	runStreamWorker(t, newMockDequeuer(job), store, exec, pub)

	if res := store.latest(); res == nil || res.status != models.InvocationCompleted {
		t.Fatalf("expected completed status, got %+v", res)
	}

	events := pub.forID(job.InvocationID)
	if got := countEventType(events, "log"); got != 0 {
		t.Errorf("expected 0 log events in prod (dev-gated), got %d (events=%v)", got, events)
	}
	if countEventType(events, "result") != 1 {
		t.Errorf("expected result event to still be forwarded in prod, got %d", countEventType(events, "result"))
	}
}

// --- Helpers ---

func makeJob(invocationID string) *redisstore.Job {
	return &redisstore.Job{
		InvocationID: invocationID,
		SnippetID:    "snip-1",
		VersionID:    "ver-1",
		TenantID:     "tenant-1",
		Language:     "bun",
		Code:         "export default async function handler() { return 42 }",
		Input:        `{}`,
		TimeoutMs:    5000,
		MaxMemoryMB:  128,
	}
}

// runWorkerUntilDone runs the worker, waits until expectedJobs results are
// persisted or 5 seconds elapse, then cancels the worker context.
func runWorkerUntilDone(t *testing.T, deq *mockDequeuer, store *mockWorkerStore, exec executor.Executor, expectedJobs int) {
	t.Helper()

	log := zap.NewNop()
	w := worker.New(deq, store, exec, log, 1)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		defer close(done)
		w.Run(ctx)
	}()

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if store.count() >= expectedJobs {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Error("worker did not stop within 2s after cancel")
	}
}

// --- Tests ---

func TestWorker_ProcessesJob(t *testing.T) {
	job := makeJob("inv-1")

	exec := &mockExecutor{
		run: func(ctx context.Context, spec executor.RunSpec) executor.RunResult {
			return executor.RunResult{Output: `{"ok":true}`, ExitCode: 0}
		},
	}
	store := &mockWorkerStore{}
	deq := newMockDequeuer(job)

	runWorkerUntilDone(t, deq, store, exec, 1)

	result := store.latest()
	if result == nil {
		t.Fatal("no result stored")
	}
	if result.id != job.InvocationID {
		t.Errorf("id = %q; want %q", result.id, job.InvocationID)
	}
	if result.status != models.InvocationCompleted {
		t.Errorf("status = %q; want %q", result.status, models.InvocationCompleted)
	}
	if result.output != `{"ok":true}` {
		t.Errorf("output = %q; want %q", result.output, `{"ok":true}`)
	}
}

func TestWorker_TimeoutMapsToTimeoutStatus(t *testing.T) {
	job := makeJob("inv-timeout")

	exec := &mockExecutor{
		run: func(ctx context.Context, spec executor.RunSpec) executor.RunResult {
			return executor.RunResult{Error: "timeout", ExitCode: -1}
		},
	}
	store := &mockWorkerStore{}
	deq := newMockDequeuer(job)

	runWorkerUntilDone(t, deq, store, exec, 1)

	result := store.latest()
	if result == nil {
		t.Fatal("no result stored")
	}
	if result.status != models.InvocationTimeout {
		t.Errorf("status = %q; want %q", result.status, models.InvocationTimeout)
	}
}

func TestWorker_OOMKilledStatus(t *testing.T) {
	job := makeJob("inv-oom")

	exec := &mockExecutor{
		run: func(ctx context.Context, spec executor.RunSpec) executor.RunResult {
			return executor.RunResult{Error: "oom", ExitCode: 137}
		},
	}
	store := &mockWorkerStore{}
	deq := newMockDequeuer(job)

	runWorkerUntilDone(t, deq, store, exec, 1)

	result := store.latest()
	if result == nil {
		t.Fatal("no result stored")
	}
	if result.status != models.InvocationOOMKilled {
		t.Errorf("status = %q; want %q", result.status, models.InvocationOOMKilled)
	}
}

func TestWorker_WebhookFiredOnCallbackURL(t *testing.T) {
	// Webhook delivery is tested in depth in webhook_test.go using httptest.
	// Here we verify the job still completes even when the callback URL is
	// unreachable (best-effort; failure is logged but not fatal).
	job := makeJob("inv-webhook")
	job.CallbackURL = "http://127.0.0.1:1/unreachable" // port 1 is not routable

	exec := &mockExecutor{
		run: func(ctx context.Context, spec executor.RunSpec) executor.RunResult {
			return executor.RunResult{Output: `{}`, ExitCode: 0}
		},
	}
	store := &mockWorkerStore{}
	deq := newMockDequeuer(job)

	runWorkerUntilDone(t, deq, store, exec, 1)

	result := store.latest()
	if result == nil {
		t.Fatal("no result stored")
	}
	if result.status != models.InvocationCompleted {
		t.Errorf("status = %q; want %q", result.status, models.InvocationCompleted)
	}
}

func TestWorker_ProcessesMultipleJobs(t *testing.T) {
	n := 5
	jobs := make([]*redisstore.Job, n)
	for i := range jobs {
		jobs[i] = makeJob(fmt.Sprintf("inv-%d", i))
	}

	exec := &mockExecutor{
		run: func(ctx context.Context, spec executor.RunSpec) executor.RunResult {
			return executor.RunResult{Output: `{}`, ExitCode: 0}
		},
	}
	store := &mockWorkerStore{}
	deq := newMockDequeuer(jobs...)

	runWorkerUntilDone(t, deq, store, exec, n)

	if got := store.count(); got != n {
		t.Errorf("processed %d jobs; want %d", got, n)
	}
}

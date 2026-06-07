// Package worker implements the background job processor that pulls async
// invocations off the Redis queue and executes them.
package worker

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/abskrj/velane/services/control-plane/internal/executor"
	"github.com/abskrj/velane/services/control-plane/internal/models"
	"github.com/abskrj/velane/services/control-plane/internal/observability"
	redisstore "github.com/abskrj/velane/services/control-plane/internal/store/redis"
)

// Dequeuer is the interface the worker uses to pull jobs from the queue.
type Dequeuer interface {
	Dequeue(ctx context.Context) (*redisstore.Job, error)
}

// WorkerStore is the DB operations the worker needs.
type WorkerStore interface {
	UpdateInvocationResult(ctx context.Context, id string, status models.InvocationStatus, output, errMsg, stderr string, durationMs, peakMemoryMB, cpuMs int) error
}

// EventPublisher publishes a typed JSON event to an invocation's live event
// stream. Implemented by *redisstore.Client.
type EventPublisher interface {
	PublishEvent(ctx context.Context, invocationID string, payload []byte) error
}

// noopPublisher discards events; used when no publisher is configured.
type noopPublisher struct{}

func (noopPublisher) PublishEvent(context.Context, string, []byte) error { return nil }

// Worker pulls jobs from the Redis queue and executes them concurrently.
type Worker struct {
	queue   Dequeuer
	store   WorkerStore
	exec    executor.Executor
	webhook *WebhookClient
	log     *zap.Logger
	workers int // concurrency level
	observe observability.InvocationObserver
	events  EventPublisher
}

// New creates a Worker. workers controls how many goroutines process jobs in
// parallel.
func New(queue Dequeuer, store WorkerStore, exec executor.Executor, log *zap.Logger, workers int) *Worker {
	if workers <= 0 {
		workers = 1
	}
	return &Worker{
		queue:   queue,
		store:   store,
		exec:    exec,
		webhook: newWebhookClient(log),
		log:     log,
		workers: workers,
		observe: observability.NoopObserver{},
		events:  noopPublisher{},
	}
}

// SetObserver injects a post-invocation observer for observability pipelines.
func (w *Worker) SetObserver(observer observability.InvocationObserver) {
	if observer == nil {
		w.observe = observability.NoopObserver{}
		return
	}
	w.observe = observer
}

// SetEventPublisher injects the live event publisher used by streaming jobs.
func (w *Worker) SetEventPublisher(p EventPublisher) {
	if p == nil {
		w.events = noopPublisher{}
		return
	}
	w.events = p
}

// Run starts w.workers goroutines, each blocking on Dequeue, and processes
// jobs until ctx is cancelled.
func (w *Worker) Run(ctx context.Context) {
	var wg sync.WaitGroup
	for i := 0; i < w.workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			w.loop(ctx, workerID)
		}(i)
	}
	wg.Wait()
}

// loop is the per-goroutine job processing loop.
func (w *Worker) loop(ctx context.Context, workerID int) {
	for {
		if ctx.Err() != nil {
			return
		}

		job, err := w.queue.Dequeue(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			w.log.Error("worker: dequeue error", zap.Int("worker", workerID), zap.Error(err))
			continue
		}

		// Dequeue returns (nil, nil) on context cancellation.
		if job == nil {
			return
		}

		w.process(ctx, workerID, job)
	}
}

// specForJob builds the executor RunSpec from a queued job.
func specForJob(job *redisstore.Job) executor.RunSpec {
	spec := executor.RunSpec{
		Language:      job.Language,
		Code:          job.Code,
		Input:         job.Input,
		TimeoutMs:     job.TimeoutMs,
		MaxMemoryMB:   job.MaxMemoryMB,
		SecretEnvVars: job.SecretEnvVars,
		Libraries:     job.Libraries,
	}
	if job.EgressPolicy != nil {
		spec.EgressPolicy = &executor.EgressPolicy{
			BlockedCIDRs:   job.EgressPolicy.BlockedCIDRs,
			BlockedDomains: job.EgressPolicy.BlockedDomains,
		}
	}
	return spec
}

// process executes a single job and persists the result, dispatching to the
// streaming or buffered path based on the job mode.
func (w *Worker) process(ctx context.Context, workerID int, job *redisstore.Job) {
	w.log.Info("worker: executing job",
		zap.Int("worker", workerID),
		zap.String("invocation_id", job.InvocationID),
		zap.String("language", job.Language),
		zap.Bool("stream", job.Stream),
	)

	if job.Stream {
		w.processStream(ctx, workerID, job)
		return
	}

	result := w.exec.Run(ctx, specForJob(job))

	status := mapResultStatus(result)

	if err := w.store.UpdateInvocationResult(ctx,
		job.InvocationID,
		status,
		result.Output,
		result.Error,
		result.Stderr,
		result.DurationMs,
		result.PeakMemoryMB,
		0,
	); err != nil {
		w.log.Error("worker: persist result",
			zap.String("invocation_id", job.InvocationID),
			zap.Error(err),
		)
	}

	_ = w.observe.OnInvocationCompleted(ctx, &models.Invocation{
		ID:           job.InvocationID,
		SnippetID:    job.SnippetID,
		VersionID:    job.VersionID,
		Environment:  job.Env,
		TenantID:     job.TenantID,
		Status:       status,
		InputPayload: job.Input,
		Output:       result.Output,
		Error:        result.Error,
		Stderr:       result.Stderr,
		DurationMs:   result.DurationMs,
		PeakMemoryMB: result.PeakMemoryMB,
		CPUMs:        0,
		CallbackURL:  job.CallbackURL,
		InvokeMode:   "async",
	})

	if job.CallbackURL != "" {
		w.webhook.Deliver(ctx, job.CallbackURL, WebhookPayload{
			InvocationID: job.InvocationID,
			Status:       string(status),
			Output:       result.Output,
			Error:        result.Error,
			DurationMs:   result.DurationMs,
		})
	}

	w.log.Info("worker: job done",
		zap.Int("worker", workerID),
		zap.String("invocation_id", job.InvocationID),
		zap.String("status", string(status)),
	)
}

// processStream executes a job via the streaming path, publishing typed events
// to the invocation's live event stream as they arrive, then finalizes the DB
// record. Debug "log" events are dropped unless the invocation env is "dev".
func (w *Worker) processStream(ctx context.Context, workerID int, job *redisstore.Job) {
	start := time.Now()
	devMode := job.Env == "dev"

	ch, err := w.exec.RunStream(ctx, specForJob(job))
	if err != nil {
		w.log.Error("worker: run stream",
			zap.String("invocation_id", job.InvocationID), zap.Error(err))
		w.publish(ctx, job.InvocationID, executor.StreamChunk{Type: "error", Message: err.Error(), ExitCode: -1})
		w.publish(ctx, job.InvocationID, executor.StreamChunk{Type: "done", Done: true})
		w.finalizeStream(ctx, job, executor.RunResult{Error: err.Error(), ExitCode: -1, DurationMs: int(time.Since(start).Milliseconds())})
		return
	}

	var (
		output   string
		stderr   string
		errMsg   string
		exitCode int
	)

	for chunk := range ch {
		ev := normalizeEvent(chunk)

		// Dev-gating: debug logs never leave the worker outside dev.
		if ev.Type == "log" && !devMode {
			continue
		}

		// The terminal "done" sentinel is deferred until after the DB is
		// finalized, so a JSON caller that stops on "done" reads a complete row.
		if ev.Done || ev.Type == "done" {
			break
		}

		switch ev.Type {
		case "log":
			if ev.Stream == "stderr" && ev.Text != "" {
				stderr += ev.Text + "\n"
			}
		case "chunk":
			output += ev.Data
		case "result":
			output = ev.Output
			if ev.ExitCode != 0 {
				exitCode = ev.ExitCode
			}
		case "error":
			if ev.Message != "" {
				errMsg = ev.Message
			} else {
				errMsg = ev.Error
			}
			if ev.ExitCode != 0 {
				exitCode = ev.ExitCode
			} else {
				exitCode = 1
			}
		}

		w.publish(ctx, job.InvocationID, ev)
	}

	// Finalize the durable record BEFORE the terminal event so that callers
	// reacting on "done" observe the committed result.
	w.finalizeStream(ctx, job, executor.RunResult{
		Output:     output,
		Stderr:     stderr,
		Error:      errMsg,
		ExitCode:   exitCode,
		DurationMs: int(time.Since(start).Milliseconds()),
	})

	// Terminal sentinel, always emitted last.
	w.publish(ctx, job.InvocationID, executor.StreamChunk{Type: "done", Done: true})

	w.log.Info("worker: stream job done",
		zap.Int("worker", workerID),
		zap.String("invocation_id", job.InvocationID),
	)
}

// publish marshals a typed event and appends it to the invocation event stream.
func (w *Worker) publish(ctx context.Context, invocationID string, ev executor.StreamChunk) {
	payload, err := json.Marshal(ev)
	if err != nil {
		return
	}
	if err := w.events.PublishEvent(ctx, invocationID, payload); err != nil {
		w.log.Debug("worker: publish event", zap.String("invocation_id", invocationID), zap.Error(err))
	}
}

// finalizeStream persists the accumulated result and fires observers / webhook,
// mirroring the buffered path's terminal behaviour.
func (w *Worker) finalizeStream(ctx context.Context, job *redisstore.Job, result executor.RunResult) {
	status := mapResultStatus(result)

	if err := w.store.UpdateInvocationResult(ctx,
		job.InvocationID, status, result.Output, result.Error, result.Stderr,
		result.DurationMs, result.PeakMemoryMB, 0,
	); err != nil {
		w.log.Error("worker: persist stream result",
			zap.String("invocation_id", job.InvocationID), zap.Error(err))
	}

	_ = w.observe.OnInvocationCompleted(ctx, &models.Invocation{
		ID:           job.InvocationID,
		SnippetID:    job.SnippetID,
		VersionID:    job.VersionID,
		Environment:  job.Env,
		TenantID:     job.TenantID,
		Status:       status,
		InputPayload: job.Input,
		Output:       result.Output,
		Error:        result.Error,
		Stderr:       result.Stderr,
		DurationMs:   result.DurationMs,
		PeakMemoryMB: result.PeakMemoryMB,
		InvokeMode:   "stream",
	})

	if job.CallbackURL != "" {
		w.webhook.Deliver(ctx, job.CallbackURL, WebhookPayload{
			InvocationID: job.InvocationID,
			Status:       string(status),
			Output:       result.Output,
			Error:        result.Error,
			DurationMs:   result.DurationMs,
		})
	}
}

// normalizeEvent fills in the Type field for legacy events that carry only
// {data, error, done} so downstream consumers can switch on Type uniformly.
func normalizeEvent(c executor.StreamChunk) executor.StreamChunk {
	if c.Type != "" {
		return c
	}
	switch {
	case c.Error != "":
		c.Type = "error"
		c.Message = c.Error
	case c.Done && c.Data != "":
		// Legacy single terminal chunk: treat its payload as the result.
		c.Type = "result"
		c.Output = c.Data
	case c.Done:
		c.Type = "done"
	default:
		c.Type = "chunk"
	}
	return c
}

// mapResultStatus converts an executor RunResult into an InvocationStatus.
func mapResultStatus(result executor.RunResult) models.InvocationStatus {
	if result.Error == "timeout" {
		return models.InvocationTimeout
	}
	if result.Error == "oom" {
		return models.InvocationOOMKilled
	}
	if result.ExitCode != 0 || result.Error != "" {
		return models.InvocationFailed
	}
	return models.InvocationCompleted
}

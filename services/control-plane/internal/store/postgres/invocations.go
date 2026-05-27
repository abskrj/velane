package postgres

import (
	"context"
	"fmt"

	"github.com/runeforge/control-plane/internal/models"
)

// CreateInvocation inserts a new invocation record with status=running.
func (s *Store) CreateInvocation(ctx context.Context, snippetID, versionID, environment, tenantID, inputPayload string) (*models.Invocation, error) {
	row := s.pool.QueryRow(ctx,
		`INSERT INTO invocations (snippet_id, version_id, environment, tenant_id, status, input_payload)
		 VALUES ($1, $2, $3, $4, 'running', $5)
		 RETURNING id, snippet_id, version_id, environment, tenant_id, status,
		           input_payload, output, error, stderr, duration_ms, peak_memory_mb,
		           created_at, completed_at`,
		snippetID, versionID, environment, tenantID, inputPayload,
	)
	inv, err := scanInvocation(row)
	if err != nil {
		return nil, fmt.Errorf("CreateInvocation scan: %w", err)
	}
	return inv, nil
}

// UpdateInvocationResult updates an invocation with the execution result.
func (s *Store) UpdateInvocationResult(ctx context.Context, id string, status models.InvocationStatus, output, errMsg, stderr string, durationMs, peakMemoryMB int) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE invocations
		 SET status       = $2,
		     output       = $3,
		     error        = $4,
		     stderr       = $5,
		     duration_ms  = $6,
		     peak_memory_mb = $7,
		     completed_at = now()
		 WHERE id = $1`,
		id, string(status), output, errMsg, stderr, durationMs, peakMemoryMB,
	)
	if err != nil {
		return fmt.Errorf("UpdateInvocationResult: %w", err)
	}
	return nil
}

// GetInvocation retrieves a single invocation by its primary key.
func (s *Store) GetInvocation(ctx context.Context, id string) (*models.Invocation, error) {
	row := s.pool.QueryRow(ctx,
		`SELECT id, snippet_id, version_id, environment, tenant_id, status,
		        input_payload, output, error, stderr, duration_ms, peak_memory_mb,
		        created_at, completed_at
		 FROM invocations WHERE id = $1`,
		id,
	)
	inv, err := scanInvocation(row)
	if err != nil {
		return nil, fmt.Errorf("GetInvocation: %w", err)
	}
	return inv, nil
}

// ListInvocationsBySnippet returns recent invocations for a snippet.
func (s *Store) ListInvocationsBySnippet(ctx context.Context, snippetID string, limit int) ([]*models.Invocation, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, snippet_id, version_id, environment, tenant_id, status,
		        input_payload, output, error, stderr, duration_ms, peak_memory_mb,
		        created_at, completed_at
		 FROM invocations WHERE snippet_id = $1
		 ORDER BY created_at DESC LIMIT $2`,
		snippetID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("ListInvocationsBySnippet query: %w", err)
	}
	defer rows.Close()

	var invocations []*models.Invocation
	for rows.Next() {
		inv, err := scanInvocation(rows)
		if err != nil {
			return nil, err
		}
		invocations = append(invocations, inv)
	}
	return invocations, rows.Err()
}

func scanInvocation(s scannable) (*models.Invocation, error) {
	var inv models.Invocation
	var output, errMsg, stderr *string
	var durationMs, peakMemoryMB *int
	if err := s.Scan(
		&inv.ID, &inv.SnippetID, &inv.VersionID, &inv.Environment, &inv.TenantID,
		&inv.Status, &inv.InputPayload,
		&output, &errMsg, &stderr,
		&durationMs, &peakMemoryMB,
		&inv.CreatedAt, &inv.CompletedAt,
	); err != nil {
		return nil, err
	}
	if output != nil {
		inv.Output = *output
	}
	if errMsg != nil {
		inv.Error = *errMsg
	}
	if stderr != nil {
		inv.Stderr = *stderr
	}
	if durationMs != nil {
		inv.DurationMs = *durationMs
	}
	if peakMemoryMB != nil {
		inv.PeakMemoryMB = *peakMemoryMB
	}
	return &inv, nil
}

package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/abskrj/velane/services/control-plane/internal/models"
)

// UpsertConnection creates or updates a connection record for a tenant + provider.
func (s *Store) UpsertConnection(ctx context.Context, tenantID, provider, nangoConnectionID, displayName string) (*models.Connection, error) {
	now := time.Now()
	row := s.pool.QueryRow(ctx,
		`INSERT INTO connections (tenant_id, provider, nango_connection_id, display_name, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $5)
		 ON CONFLICT (tenant_id, provider) DO UPDATE
		   SET nango_connection_id = EXCLUDED.nango_connection_id,
		       display_name        = EXCLUDED.display_name,
		       updated_at          = EXCLUDED.updated_at
		 RETURNING id, tenant_id, provider, nango_connection_id, display_name, created_at, updated_at`,
		tenantID, provider, nangoConnectionID, displayName, now,
	)
	var c models.Connection
	if err := row.Scan(&c.ID, &c.TenantID, &c.Provider, &c.NangoConnectionID, &c.DisplayName, &c.CreatedAt, &c.UpdatedAt); err != nil {
		return nil, fmt.Errorf("UpsertConnection: %w", err)
	}
	return &c, nil
}

// ListConnections returns all connections for a tenant.
func (s *Store) ListConnections(ctx context.Context, tenantID string) ([]*models.Connection, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, tenant_id, provider, nango_connection_id, display_name, created_at, updated_at
		 FROM connections WHERE tenant_id = $1 ORDER BY provider ASC`,
		tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("ListConnections: %w", err)
	}
	defer rows.Close()

	var conns []*models.Connection
	for rows.Next() {
		var c models.Connection
		if err := rows.Scan(&c.ID, &c.TenantID, &c.Provider, &c.NangoConnectionID, &c.DisplayName, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		conns = append(conns, &c)
	}
	return conns, rows.Err()
}

// GetConnection returns a single connection by tenant + provider.
func (s *Store) GetConnection(ctx context.Context, tenantID, provider string) (*models.Connection, error) {
	row := s.pool.QueryRow(ctx,
		`SELECT id, tenant_id, provider, nango_connection_id, display_name, created_at, updated_at
		 FROM connections WHERE tenant_id = $1 AND provider = $2`,
		tenantID, provider,
	)
	var c models.Connection
	if err := row.Scan(&c.ID, &c.TenantID, &c.Provider, &c.NangoConnectionID, &c.DisplayName, &c.CreatedAt, &c.UpdatedAt); err != nil {
		return nil, fmt.Errorf("GetConnection: %w", err)
	}
	return &c, nil
}

// DeleteConnection removes a connection by tenant + provider.
func (s *Store) DeleteConnection(ctx context.Context, tenantID, provider string) error {
	result, err := s.pool.Exec(ctx,
		`DELETE FROM connections WHERE tenant_id = $1 AND provider = $2`,
		tenantID, provider,
	)
	if err != nil {
		return fmt.Errorf("DeleteConnection: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("connection not found")
	}
	return nil
}

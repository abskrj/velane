package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/abskrj/velane/services/control-plane/internal/models"
)

// UpsertConnection creates or updates a connection record for (tenant, provider, alias).
// It owns the display_name field only — it never overwrites nango_connection_id so that
// a concurrent webhook delivery is not clobbered.
func (s *Store) UpsertConnection(ctx context.Context, tenantID, provider, alias, providerConfigKey string, credentialProfileID *string, displayName string) (*models.Connection, error) {
	now := time.Now()
	row := s.pool.QueryRow(ctx,
		`INSERT INTO connections (tenant_id, provider, alias, provider_config_key, credential_profile_id, nango_connection_id, display_name, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, NULL, $6, $7, $7)
		 ON CONFLICT (tenant_id, provider, alias) DO UPDATE
		   SET display_name = EXCLUDED.display_name,
		       provider_config_key = EXCLUDED.provider_config_key,
		       credential_profile_id = EXCLUDED.credential_profile_id,
		       updated_at   = EXCLUDED.updated_at
		 RETURNING id, tenant_id, provider, alias, provider_config_key, credential_profile_id,
		           COALESCE(nango_connection_id, ''), display_name, created_at, updated_at`,
		tenantID, provider, alias, providerConfigKey, credentialProfileID, displayName, now,
	)
	var c models.Connection
	if err := row.Scan(&c.ID, &c.TenantID, &c.Provider, &c.Alias, &c.ProviderConfigKey, &c.CredentialProfileID,
		&c.NangoConnectionID, &c.DisplayName, &c.CreatedAt, &c.UpdatedAt); err != nil {
		return nil, fmt.Errorf("UpsertConnection: %w", err)
	}
	return &c, nil
}

// UpdateNangoConnectionID sets the real Nango-generated UUID on an existing connection row.
// Called exclusively by the Nango webhook handler; it owns nango_connection_id and never
// touches display_name so it does not clobber the field set by UpsertConnection.
func (s *Store) UpdateNangoConnectionID(ctx context.Context, tenantID, provider, alias, nangoConnectionID string) (*models.Connection, error) {
	now := time.Now()
	row := s.pool.QueryRow(ctx,
		`UPDATE connections
		 SET nango_connection_id = $4,
		     updated_at          = $5
		 WHERE tenant_id = $1 AND provider = $2 AND alias = $3
		 RETURNING id, tenant_id, provider, alias, provider_config_key, credential_profile_id,
		           nango_connection_id, display_name, created_at, updated_at`,
		tenantID, provider, alias, nangoConnectionID, now,
	)
	var c models.Connection
	if err := row.Scan(&c.ID, &c.TenantID, &c.Provider, &c.Alias, &c.ProviderConfigKey, &c.CredentialProfileID,
		&c.NangoConnectionID, &c.DisplayName, &c.CreatedAt, &c.UpdatedAt); err != nil {
		return nil, fmt.Errorf("UpdateNangoConnectionID: %w", err)
	}
	return &c, nil
}

func (s *Store) UpdateNangoConnectionIDByProviderConfigKey(ctx context.Context, tenantID, providerConfigKey, nangoConnectionID string) (*models.Connection, error) {
	now := time.Now()
	row := s.pool.QueryRow(ctx,
		`UPDATE connections
		 SET nango_connection_id = $3,
		     updated_at          = $4
		 WHERE tenant_id = $1 AND provider_config_key = $2
		 RETURNING id, tenant_id, provider, alias, provider_config_key, credential_profile_id,
		           nango_connection_id, display_name, created_at, updated_at`,
		tenantID, providerConfigKey, nangoConnectionID, now,
	)
	var c models.Connection
	if err := row.Scan(&c.ID, &c.TenantID, &c.Provider, &c.Alias, &c.ProviderConfigKey, &c.CredentialProfileID,
		&c.NangoConnectionID, &c.DisplayName, &c.CreatedAt, &c.UpdatedAt); err != nil {
		return nil, fmt.Errorf("UpdateNangoConnectionIDByProviderConfigKey: %w", err)
	}
	return &c, nil
}

// ListConnections returns all connections for a tenant ordered by provider and alias.
func (s *Store) ListConnections(ctx context.Context, tenantID string) ([]*models.Connection, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, tenant_id, provider, alias, provider_config_key, credential_profile_id,
		        COALESCE(nango_connection_id, ''), display_name, created_at, updated_at
		 FROM connections WHERE tenant_id = $1 ORDER BY provider ASC, alias ASC`,
		tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("ListConnections: %w", err)
	}
	defer rows.Close()

	var conns []*models.Connection
	for rows.Next() {
		var c models.Connection
		if err := rows.Scan(&c.ID, &c.TenantID, &c.Provider, &c.Alias, &c.ProviderConfigKey, &c.CredentialProfileID,
			&c.NangoConnectionID, &c.DisplayName, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		conns = append(conns, &c)
	}
	return conns, rows.Err()
}

// GetConnection returns the connection for (tenant, provider) with alias 'default'.
// For targeted alias lookups use GetConnectionByAlias.
func (s *Store) GetConnection(ctx context.Context, tenantID, provider string) (*models.Connection, error) {
	return s.GetConnectionByAlias(ctx, tenantID, provider, "default")
}

// GetConnectionByAlias returns a specific connection by (tenant, provider, alias).
func (s *Store) GetConnectionByAlias(ctx context.Context, tenantID, provider, alias string) (*models.Connection, error) {
	row := s.pool.QueryRow(ctx,
		`SELECT id, tenant_id, provider, alias, provider_config_key, credential_profile_id,
		        COALESCE(nango_connection_id, ''), display_name, created_at, updated_at
		 FROM connections WHERE tenant_id = $1 AND provider = $2 AND alias = $3`,
		tenantID, provider, alias,
	)
	var c models.Connection
	if err := row.Scan(&c.ID, &c.TenantID, &c.Provider, &c.Alias, &c.ProviderConfigKey, &c.CredentialProfileID,
		&c.NangoConnectionID, &c.DisplayName, &c.CreatedAt, &c.UpdatedAt); err != nil {
		return nil, fmt.Errorf("GetConnectionByAlias: %w", err)
	}
	return &c, nil
}

// DeleteConnection removes all connections for (tenant, provider) across all aliases.
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

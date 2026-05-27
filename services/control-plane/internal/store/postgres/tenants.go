package postgres

import (
	"context"
	"fmt"

	"github.com/runeforge/control-plane/internal/models"
)

// CreateTenant inserts a new tenant row and returns the created record.
func (s *Store) CreateTenant(ctx context.Context, name, slug string) (*models.Tenant, error) {
	row := s.pool.QueryRow(ctx,
		`INSERT INTO tenants (name, slug)
		 VALUES ($1, $2)
		 RETURNING id, name, slug, created_at`,
		name, slug,
	)

	var t models.Tenant
	if err := row.Scan(&t.ID, &t.Name, &t.Slug, &t.CreatedAt); err != nil {
		return nil, fmt.Errorf("CreateTenant scan: %w", err)
	}
	return &t, nil
}

// GetTenantByID retrieves a tenant by its primary key.
func (s *Store) GetTenantByID(ctx context.Context, id string) (*models.Tenant, error) {
	row := s.pool.QueryRow(ctx,
		`SELECT id, name, slug, created_at FROM tenants WHERE id = $1`,
		id,
	)

	var t models.Tenant
	if err := row.Scan(&t.ID, &t.Name, &t.Slug, &t.CreatedAt); err != nil {
		return nil, fmt.Errorf("GetTenantByID scan: %w", err)
	}
	return &t, nil
}

// GetTenantBySlug retrieves a tenant by its unique URL slug.
func (s *Store) GetTenantBySlug(ctx context.Context, slug string) (*models.Tenant, error) {
	row := s.pool.QueryRow(ctx,
		`SELECT id, name, slug, created_at FROM tenants WHERE slug = $1`,
		slug,
	)

	var t models.Tenant
	if err := row.Scan(&t.ID, &t.Name, &t.Slug, &t.CreatedAt); err != nil {
		return nil, fmt.Errorf("GetTenantBySlug scan: %w", err)
	}
	return &t, nil
}

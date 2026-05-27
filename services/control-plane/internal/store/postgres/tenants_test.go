package postgres_test

import (
	"context"
	"strings"
	"testing"
)

func TestCreateTenant(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()

	slug := uniqueSlug(t, "acme")
	tenant, err := store.CreateTenant(ctx, "Acme Corp", slug)
	if err != nil {
		t.Fatalf("CreateTenant: %v", err)
	}

	if tenant.ID == "" {
		t.Error("tenant ID must not be empty")
	}
	if tenant.Name != "Acme Corp" {
		t.Errorf("name = %q; want %q", tenant.Name, "Acme Corp")
	}
	if tenant.Slug != slug {
		t.Errorf("slug = %q; want %q", tenant.Slug, slug)
	}
	if tenant.CreatedAt.IsZero() {
		t.Error("created_at must be set")
	}
}

func TestGetTenantByID(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()

	created, err := store.CreateTenant(ctx, "Test Org", uniqueSlug(t, "test-org"))
	if err != nil {
		t.Fatalf("CreateTenant: %v", err)
	}

	fetched, err := store.GetTenantByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetTenantByID: %v", err)
	}
	if fetched.ID != created.ID {
		t.Errorf("ID mismatch: got %q, want %q", fetched.ID, created.ID)
	}
}

func TestGetTenantBySlug(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()

	slug := uniqueSlug(t, "slug-test")
	created, err := store.CreateTenant(ctx, "Slug Test", slug)
	if err != nil {
		t.Fatalf("CreateTenant: %v", err)
	}

	fetched, err := store.GetTenantBySlug(ctx, slug)
	if err != nil {
		t.Fatalf("GetTenantBySlug: %v", err)
	}
	if fetched.ID != created.ID {
		t.Errorf("ID mismatch: got %q, want %q", fetched.ID, created.ID)
	}
}

func TestCreateTenant_DuplicateSlugFails(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()

	slug := uniqueSlug(t, "dup")
	_, err := store.CreateTenant(ctx, "First", slug)
	if err != nil {
		t.Fatalf("first CreateTenant: %v", err)
	}

	_, err = store.CreateTenant(ctx, "Second", slug)
	if err == nil {
		t.Fatal("expected error for duplicate slug, got nil")
	}
	if !strings.Contains(err.Error(), "unique") && !strings.Contains(err.Error(), "duplicate") &&
		!strings.Contains(err.Error(), "23505") {
		t.Errorf("expected unique constraint error, got: %v", err)
	}
}

func TestGetTenantByID_NotFound(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()

	_, err := store.GetTenantByID(ctx, "00000000-0000-0000-0000-000000000000")
	if err == nil {
		t.Fatal("expected error for non-existent tenant, got nil")
	}
}

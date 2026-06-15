package postgres_test

import (
	"context"
	"testing"
)

func TestCreateSnippet(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()

	tenant, err := store.CreateTenant(ctx, "Snippet Org", uniqueSlug(t, "snippet-org"))
	if err != nil {
		t.Fatalf("CreateTenant: %v", err)
	}

	sn, err := store.CreateSnippet(ctx, tenant.ID, "My Snippet", "bun", "user-1")
	if err != nil {
		t.Fatalf("CreateSnippet: %v", err)
	}

	if sn.ID == "" {
		t.Error("snippet ID must not be empty")
	}
	if sn.TenantID != tenant.ID {
		t.Errorf("tenant_id = %q; want %q", sn.TenantID, tenant.ID)
	}
	if sn.Slug != sn.ID {
		t.Errorf("slug = %q; want %q (slug must equal id)", sn.Slug, sn.ID)
	}
	if string(sn.Language) != "bun" {
		t.Errorf("language = %q; want %q", sn.Language, "bun")
	}
	if sn.CreatedAt.IsZero() {
		t.Error("created_at must be set")
	}
}

func TestGetSnippetByID(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()

	tenant, _ := store.CreateTenant(ctx, "GetByID Org", uniqueSlug(t, "getbyid-org"))
	created, err := store.CreateSnippet(ctx, tenant.ID, "By ID", "python", "user-1")
	if err != nil {
		t.Fatalf("CreateSnippet: %v", err)
	}

	fetched, err := store.GetSnippetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetSnippetByID: %v", err)
	}
	if fetched.ID != created.ID {
		t.Errorf("ID mismatch: got %q, want %q", fetched.ID, created.ID)
	}
}

func TestGetSnippetBySlug(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()

	tenant, _ := store.CreateTenant(ctx, "BySlug Org", uniqueSlug(t, "byslug-org"))
	created, err := store.CreateSnippet(ctx, tenant.ID, "By Slug", "bun", "user-1")
	if err != nil {
		t.Fatalf("CreateSnippet: %v", err)
	}

	fetched, err := store.GetSnippetBySlug(ctx, tenant.ID, created.ID)
	if err != nil {
		t.Fatalf("GetSnippetBySlug: %v", err)
	}
	if fetched.ID != created.ID {
		t.Errorf("ID mismatch: got %q, want %q", fetched.ID, created.ID)
	}
}

func TestListSnippets(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()

	tenant, _ := store.CreateTenant(ctx, "List Org", uniqueSlug(t, "list-org"))

	for i := 0; i < 3; i++ {
		_, err := store.CreateSnippet(ctx, tenant.ID, "Snippet", "bun", "user-1")
		if err != nil {
			t.Fatalf("CreateSnippet %d: %v", i, err)
		}
	}

	snippets, err := store.ListSnippets(ctx, tenant.ID)
	if err != nil {
		t.Fatalf("ListSnippets: %v", err)
	}
	if len(snippets) < 3 {
		t.Errorf("expected at least 3 snippets, got %d", len(snippets))
	}
}

func TestDeleteSnippet(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()

	tenant, _ := store.CreateTenant(ctx, "Delete Org", uniqueSlug(t, "delete-org"))
	sn, err := store.CreateSnippet(ctx, tenant.ID, "To Delete", "bun", "user-1")
	if err != nil {
		t.Fatalf("CreateSnippet: %v", err)
	}

	if err := store.DeleteSnippet(ctx, sn.ID); err != nil {
		t.Fatalf("DeleteSnippet: %v", err)
	}

	_, err = store.GetSnippetByID(ctx, sn.ID)
	if err == nil {
		t.Error("expected error after deletion, got nil")
	}
}

func TestGetSnippetEnvironment_SeedOnCreate(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()

	tenant, _ := store.CreateTenant(ctx, "Env Org", uniqueSlug(t, "env-org"))
	sn, err := store.CreateSnippet(ctx, tenant.ID, "Env Snippet", "python", "user-1")
	if err != nil {
		t.Fatalf("CreateSnippet: %v", err)
	}

	for _, env := range []string{"dev", "prod"} {
		se, err := store.GetSnippetEnvironment(ctx, sn.ID, env)
		if err != nil {
			t.Fatalf("GetSnippetEnvironment(%q): %v", env, err)
		}
		if se.ActiveVersionID != nil {
			t.Errorf("env %q: expected nil active version on fresh snippet, got %q", env, *se.ActiveVersionID)
		}
	}
}

func TestCreateSnippet_SlugEqualsID(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()

	tenant, _ := store.CreateTenant(ctx, "Slug ID Org", uniqueSlug(t, "slugid-org"))

	first, err := store.CreateSnippet(ctx, tenant.ID, "First", "bun", "user-1")
	if err != nil {
		t.Fatalf("first CreateSnippet: %v", err)
	}
	second, err := store.CreateSnippet(ctx, tenant.ID, "Second", "bun", "user-1")
	if err != nil {
		t.Fatalf("second CreateSnippet: %v", err)
	}
	if first.Slug != first.ID || second.Slug != second.ID {
		t.Fatalf("slug must equal id: first=%q/%q second=%q/%q", first.ID, first.Slug, second.ID, second.Slug)
	}
	if first.ID == second.ID {
		t.Fatal("expected distinct workflow IDs")
	}
}

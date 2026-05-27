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

	slug := uniqueSlug(t, "my-snippet")
	sn, err := store.CreateSnippet(ctx, tenant.ID, "My Snippet", slug, "bun", "user-1")
	if err != nil {
		t.Fatalf("CreateSnippet: %v", err)
	}

	if sn.ID == "" {
		t.Error("snippet ID must not be empty")
	}
	if sn.TenantID != tenant.ID {
		t.Errorf("tenant_id = %q; want %q", sn.TenantID, tenant.ID)
	}
	if sn.Slug != slug {
		t.Errorf("slug = %q; want %q", sn.Slug, slug)
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
	created, err := store.CreateSnippet(ctx, tenant.ID, "By ID", uniqueSlug(t, "byid"), "python", "user-1")
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
	slug := uniqueSlug(t, "by-slug")
	created, err := store.CreateSnippet(ctx, tenant.ID, "By Slug", slug, "bun", "user-1")
	if err != nil {
		t.Fatalf("CreateSnippet: %v", err)
	}

	fetched, err := store.GetSnippetBySlug(ctx, tenant.ID, slug)
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
		_, err := store.CreateSnippet(ctx, tenant.ID,
			"Snippet", uniqueSlug(t, "list-snippet"), "bun", "user-1")
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
	sn, err := store.CreateSnippet(ctx, tenant.ID, "To Delete", uniqueSlug(t, "to-delete"), "bun", "user-1")
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
	sn, err := store.CreateSnippet(ctx, tenant.ID, "Env Snippet", uniqueSlug(t, "env-sn"), "python", "user-1")
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

func TestCreateSnippet_DuplicateSlugInSameTenantFails(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()

	tenant, _ := store.CreateTenant(ctx, "Dup Slug Org", uniqueSlug(t, "dupslug-org"))
	slug := uniqueSlug(t, "dup-slug")

	_, err := store.CreateSnippet(ctx, tenant.ID, "First", slug, "bun", "user-1")
	if err != nil {
		t.Fatalf("first CreateSnippet: %v", err)
	}

	_, err = store.CreateSnippet(ctx, tenant.ID, "Second", slug, "bun", "user-1")
	if err == nil {
		t.Fatal("expected error for duplicate slug within same tenant, got nil")
	}
}

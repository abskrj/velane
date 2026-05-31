// Package api_test — integration tests for connections, proxy and integrations APIs.
// These tests require TEST_DATABASE_URL to be set; they are skipped otherwise.
package api_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	api "github.com/abskrj/velane/services/control-plane/internal/api"
	"github.com/abskrj/velane/services/control-plane/internal/auth"
	"github.com/abskrj/velane/services/control-plane/internal/executor/remote"
	"github.com/abskrj/velane/services/control-plane/internal/nango"
	"github.com/abskrj/velane/services/control-plane/internal/scheduler"
	"github.com/abskrj/velane/services/control-plane/internal/store/postgres"
	"go.uber.org/zap"
)

// setupWithNango wires the full stack with a mock Nango server.
func setupWithNango(t *testing.T) *testEnv {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set — skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	store, err := postgres.New(ctx, dsn)
	if err != nil {
		t.Fatalf("connect test postgres: %v", err)
	}
	t.Cleanup(func() { store.Close() })

	// Mock Nango server handling the subset of endpoints we need.
	mockNango := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		// POST /connect/sessions — returns a session token.
		case r.Method == http.MethodPost && r.URL.Path == "/connect/sessions":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":{"token":"mock-session-token"}}`))

		// DELETE /connection/{anything} — 204 no body.
		case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/connection/"):
			w.WriteHeader(http.StatusNoContent)

		// GET /providers — returns a minimal provider list.
		case r.Method == http.MethodGet && r.URL.Path == "/providers":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[` +
				`{"unique_key":"github","name":"GitHub","auth_mode":"OAUTH2","categories":["developer-tools"]},` +
				`{"unique_key":"slack","name":"Slack","auth_mode":"OAUTH2","categories":["communication"]},` +
				`{"unique_key":"figma","name":"Figma","auth_mode":"OAUTH2","categories":["design"]}` +
				`]`))

		// GET /proxy/* — forward as a JSON response from the mock provider.
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/proxy/"):
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"login":"mockuser"}`))

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(mockNango.Close)

	nangoClient := nango.New(mockNango.URL, "test-key")

	// Mock executor: always succeeds.
	mockExec := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"output":"{\"ok\":true}","stderr":"","duration_ms":10,"peak_memory_mb":8,"exit_code":0,"error":""}`))
	}))
	t.Cleanup(mockExec.Close)

	exec := remote.New(mockExec.URL, mockExec.URL)
	sched := scheduler.New(store, exec, testEncKey, nil)
	log := zap.NewNop()
	router := api.NewRouter(store, sched, log, testEncKey, auth.NewPasswordProvider(store), nangoClient, "", "", "", "", nil)

	slug := fmt.Sprintf("test-nango-%d", time.Now().UnixNano())
	tenant, err := store.CreateTenant(context.Background(), "Nango Tenant", slug)
	if err != nil {
		t.Fatalf("create test tenant: %v", err)
	}

	_, manageKey, err := store.CreateAPIKeyWithPlain(context.Background(), tenant.ID, "manage-key",
		[]string{"admin", "manage", "invoke"})
	if err != nil {
		t.Fatalf("create manage key: %v", err)
	}

	_, invokeKey, err := store.CreateAPIKeyWithPlain(context.Background(), tenant.ID, "invoke-key",
		[]string{"invoke"})
	if err != nil {
		t.Fatalf("create invoke key: %v", err)
	}

	return &testEnv{
		router:    router,
		store:     store,
		tenant:    tenant,
		manageKey: manageKey,
		invokeKey: invokeKey,
	}
}

// --- Connections CRUD ---

func TestConnections_ListEmpty(t *testing.T) {
	env := setupWithNango(t)
	path := fmt.Sprintf("/v1/tenants/%s/connections", env.tenant.Slug)
	rec := env.do(t, http.MethodGet, path, env.manageKey, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200\nbody: %s", rec.Code, rec.Body.String())
	}
	var conns []any
	if err := json.NewDecoder(rec.Body).Decode(&conns); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(conns) != 0 {
		t.Errorf("expected empty list, got %d connections", len(conns))
	}
}

func TestConnections_Record(t *testing.T) {
	env := setupWithNango(t)
	path := fmt.Sprintf("/v1/tenants/%s/connections", env.tenant.Slug)
	rec := env.do(t, http.MethodPost, path, env.manageKey, map[string]any{
		"provider": "github",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d; want 201\nbody: %s", rec.Code, rec.Body.String())
	}
	body := decodeJSON(t, rec)
	if body["provider"] != "github" {
		t.Errorf("provider = %q; want %q", body["provider"], "github")
	}
	if body["id"] == "" {
		t.Error("id must be set")
	}
}

func TestConnections_ListAfterRecord(t *testing.T) {
	env := setupWithNango(t)
	path := fmt.Sprintf("/v1/tenants/%s/connections", env.tenant.Slug)

	// Record a connection first.
	rec := env.do(t, http.MethodPost, path, env.manageKey, map[string]any{
		"provider": "github",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("record connection: status = %d; want 201\nbody: %s", rec.Code, rec.Body.String())
	}

	// List should show it.
	recList := env.do(t, http.MethodGet, path, env.manageKey, nil)
	if recList.Code != http.StatusOK {
		t.Fatalf("list status = %d; want 200", recList.Code)
	}
	var conns []map[string]any
	if err := json.NewDecoder(recList.Body).Decode(&conns); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(conns) == 0 {
		t.Fatal("expected at least one connection after recording")
	}
	found := false
	for _, c := range conns {
		if c["provider"] == "github" {
			found = true
			break
		}
	}
	if !found {
		t.Error("github connection not found in list after recording")
	}
}

func TestConnections_RecordRequiresManageScope(t *testing.T) {
	env := setupWithNango(t)
	path := fmt.Sprintf("/v1/tenants/%s/connections", env.tenant.Slug)
	rec := env.do(t, http.MethodPost, path, env.invokeKey, map[string]any{
		"provider": "github",
	})
	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d; want 403", rec.Code)
	}
}

func TestConnections_TenantIsolation(t *testing.T) {
	env := setupWithNango(t)
	path := fmt.Sprintf("/v1/tenants/%s/connections", env.tenant.Slug)

	// Record a connection under env's tenant.
	rec := env.do(t, http.MethodPost, path, env.manageKey, map[string]any{
		"provider": "github",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("record connection: status = %d; want 201\nbody: %s", rec.Code, rec.Body.String())
	}

	// Create a second tenant and key.
	slug2 := fmt.Sprintf("nango-t2-%d", time.Now().UnixNano())
	tenant2, err := env.store.CreateTenant(context.Background(), "T2", slug2)
	if err != nil {
		t.Fatalf("create tenant2: %v", err)
	}
	_, key2, err := env.store.CreateAPIKeyWithPlain(context.Background(), tenant2.ID, "k2",
		[]string{"invoke", "manage"})
	if err != nil {
		t.Fatalf("create key2: %v", err)
	}

	// Tenant 2's key should NOT be able to list tenant 1's connections
	// (the slug in the path belongs to tenant 1, but auth token is for tenant 2).
	rec2 := env.do(t, http.MethodGet, path, key2, nil)
	if rec2.Code != http.StatusForbidden {
		t.Errorf("cross-tenant list: status = %d; want 403", rec2.Code)
	}
}

// --- Connection session ---

func TestConnections_CreateSession(t *testing.T) {
	env := setupWithNango(t)
	path := fmt.Sprintf("/v1/tenants/%s/connections/session", env.tenant.Slug)
	rec := env.do(t, http.MethodPost, path, env.manageKey, map[string]any{
		"provider": "github",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200\nbody: %s", rec.Code, rec.Body.String())
	}
	body := decodeJSON(t, rec)
	if body["session_token"] != "mock-session-token" {
		t.Errorf("session_token = %q; want %q", body["session_token"], "mock-session-token")
	}
}

func TestConnections_CreateSessionRequiresManageScope(t *testing.T) {
	env := setupWithNango(t)
	path := fmt.Sprintf("/v1/tenants/%s/connections/session", env.tenant.Slug)
	rec := env.do(t, http.MethodPost, path, env.invokeKey, map[string]any{
		"provider": "github",
	})
	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d; want 403", rec.Code)
	}
}

func TestConnections_CreateSessionMissingProvider(t *testing.T) {
	env := setupWithNango(t)
	path := fmt.Sprintf("/v1/tenants/%s/connections/session", env.tenant.Slug)
	rec := env.do(t, http.MethodPost, path, env.manageKey, map[string]any{})
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want 400", rec.Code)
	}
}

// --- Disconnect ---

func TestConnections_Disconnect(t *testing.T) {
	env := setupWithNango(t)
	basePath := fmt.Sprintf("/v1/tenants/%s/connections", env.tenant.Slug)

	// Record a connection first.
	rec := env.do(t, http.MethodPost, basePath, env.manageKey, map[string]any{
		"provider": "github",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("record connection: status = %d; want 201\nbody: %s", rec.Code, rec.Body.String())
	}

	// Disconnect it.
	delPath := basePath + "/github"
	recDel := env.do(t, http.MethodDelete, delPath, env.manageKey, nil)
	if recDel.Code != http.StatusNoContent {
		t.Fatalf("disconnect: status = %d; want 204\nbody: %s", recDel.Code, recDel.Body.String())
	}

	// List should now be empty.
	recList := env.do(t, http.MethodGet, basePath, env.manageKey, nil)
	if recList.Code != http.StatusOK {
		t.Fatalf("list after disconnect: status = %d; want 200", recList.Code)
	}
	var conns []any
	if err := json.NewDecoder(recList.Body).Decode(&conns); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(conns) != 0 {
		t.Errorf("expected empty list after disconnect, got %d connections", len(conns))
	}
}

func TestConnections_DisconnectNotFound(t *testing.T) {
	env := setupWithNango(t)
	path := fmt.Sprintf("/v1/tenants/%s/connections/nonexistent-provider-xyz", env.tenant.Slug)
	rec := env.do(t, http.MethodDelete, path, env.manageKey, nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d; want 404", rec.Code)
	}
}

func TestConnections_DisconnectRequiresManageScope(t *testing.T) {
	env := setupWithNango(t)
	path := fmt.Sprintf("/v1/tenants/%s/connections/github", env.tenant.Slug)
	rec := env.do(t, http.MethodDelete, path, env.invokeKey, nil)
	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d; want 403", rec.Code)
	}
}

// --- MCP shortcut ---

func TestConnections_ListForToken(t *testing.T) {
	env := setupWithNango(t)
	basePath := fmt.Sprintf("/v1/tenants/%s/connections", env.tenant.Slug)

	// Record a connection so the list is non-empty.
	rec := env.do(t, http.MethodPost, basePath, env.manageKey, map[string]any{
		"provider": "github",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("record connection: status = %d; want 201\nbody: %s", rec.Code, rec.Body.String())
	}

	// Hit the slug-free shortcut.
	recList := env.do(t, http.MethodGet, "/v1/connections", env.manageKey, nil)
	if recList.Code != http.StatusOK {
		t.Fatalf("list for token: status = %d; want 200\nbody: %s", recList.Code, recList.Body.String())
	}
	var conns []map[string]any
	if err := json.NewDecoder(recList.Body).Decode(&conns); err != nil {
		t.Fatalf("decode: %v", err)
	}
	found := false
	for _, c := range conns {
		if c["provider"] == "github" {
			found = true
			break
		}
	}
	if !found {
		t.Error("github connection not found in /v1/connections response")
	}
}

// --- Proxy ---

func TestProxy_NoConnection(t *testing.T) {
	env := setupWithNango(t)

	// No connection recorded for this tenant — proxy should return 400.
	req := httptest.NewRequest(http.MethodGet, "/v1/proxy/github/user", nil)
	req.Header.Set("X-Velane-Tenant", env.tenant.ID)
	rec := httptest.NewRecorder()
	env.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want 400", rec.Code)
	}
}

func TestProxy_ForwardsToNango(t *testing.T) {
	env := setupWithNango(t)
	basePath := fmt.Sprintf("/v1/tenants/%s/connections", env.tenant.Slug)

	// Record a github connection via the API.
	rec := env.do(t, http.MethodPost, basePath, env.manageKey, map[string]any{
		"provider": "github",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("record connection: status = %d; want 201\nbody: %s", rec.Code, rec.Body.String())
	}

	// Proxy a request — no auth header needed, identified by X-Velane-Tenant.
	req := httptest.NewRequest(http.MethodGet, "/v1/proxy/github/user", nil)
	req.Header.Set("X-Velane-Tenant", env.tenant.ID)
	recProxy := httptest.NewRecorder()
	env.router.ServeHTTP(recProxy, req)

	if recProxy.Code != http.StatusOK {
		t.Fatalf("proxy status = %d; want 200\nbody: %s", recProxy.Code, recProxy.Body.String())
	}

	var body map[string]any
	if err := json.NewDecoder(recProxy.Body).Decode(&body); err != nil {
		t.Fatalf("decode proxy response: %v", err)
	}
	if body["login"] != "mockuser" {
		t.Errorf("login = %q; want %q", body["login"], "mockuser")
	}
}

func TestProxy_MissingTenantHeader(t *testing.T) {
	env := setupWithNango(t)

	// No X-Velane-Tenant header — should return 400.
	req := httptest.NewRequest(http.MethodGet, "/v1/proxy/github/user", nil)
	rec := httptest.NewRecorder()
	env.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want 400", rec.Code)
	}
}

// --- Integrations catalog ---

func TestIntegrations_ListProviders(t *testing.T) {
	env := setupWithNango(t)

	// No auth required for the catalog endpoint.
	rec := env.do(t, http.MethodGet, "/v1/integrations", "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200\nbody: %s", rec.Code, rec.Body.String())
	}

	var providers []map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&providers); err != nil {
		t.Fatalf("decode: %v", err)
	}

	found := false
	for _, p := range providers {
		if p["unique_key"] == "github" {
			found = true
			break
		}
	}
	if !found {
		t.Error("github not found in provider list")
	}
}

func TestIntegrations_GetBundledDocs(t *testing.T) {
	env := setupWithNango(t)

	// github is in bundled metadata — should return rich docs.
	rec := env.do(t, http.MethodGet, "/v1/integrations/github/docs", env.manageKey, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200\nbody: %s", rec.Code, rec.Body.String())
	}

	body := decodeJSON(t, rec)
	if body["provider"] != "github" {
		t.Errorf("provider = %q; want %q", body["provider"], "github")
	}

	endpoints, ok := body["common_endpoints"].([]any)
	if !ok {
		t.Fatalf("common_endpoints missing or wrong type: %#v", body["common_endpoints"])
	}
	if len(endpoints) == 0 {
		t.Error("expected non-empty common_endpoints for bundled github docs")
	}

	if body["bun_example"] == nil || body["bun_example"] == "" {
		t.Error("bun_example must be set in bundled docs")
	}
}

func TestIntegrations_GetFallbackDocs(t *testing.T) {
	env := setupWithNango(t)

	// figma is in the mock Nango /providers list but NOT in bundled metadata
	// → should fall through to Nango and return a valid doc structure.
	rec := env.do(t, http.MethodGet, "/v1/integrations/figma/docs", env.manageKey, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200\nbody: %s", rec.Code, rec.Body.String())
	}

	body := decodeJSON(t, rec)
	if body["provider"] != "figma" {
		t.Errorf("provider = %q; want %q", body["provider"], "figma")
	}
	// Fallback docs should still have the required fields.
	if _, ok := body["common_endpoints"]; !ok {
		t.Error("common_endpoints field missing from fallback docs")
	}
	if body["bun_example"] == nil || body["bun_example"] == "" {
		t.Error("bun_example must be set even in fallback docs")
	}
}

func TestIntegrations_UnknownProvider(t *testing.T) {
	env := setupWithNango(t)

	// Provider not in bundled metadata and not in mock Nango → 404.
	rec := env.do(t, http.MethodGet, "/v1/integrations/nonexistent-xyz/docs", env.manageKey, nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d; want 404", rec.Code)
	}
}

func TestIntegrations_DocsRequiresInvokeScope(t *testing.T) {
	env := setupWithNango(t)

	// No auth key → 401.
	rec := env.do(t, http.MethodGet, "/v1/integrations/github/docs", "", nil)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d; want 401", rec.Code)
	}
}


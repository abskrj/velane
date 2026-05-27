package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/runeforge/control-plane/internal/store/postgres"
	"go.uber.org/zap"
)

// TenantsHandler bundles all tenant-related HTTP handlers.
type TenantsHandler struct {
	store *postgres.Store
	log   *zap.Logger
}

// NewTenantsHandler constructs a TenantsHandler.
func NewTenantsHandler(store *postgres.Store, log *zap.Logger) *TenantsHandler {
	return &TenantsHandler{store: store, log: log}
}

// createTenantRequest is the expected POST body.
type createTenantRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// CreateTenant handles POST /v1/tenants.
//
// NOTE: This endpoint is intentionally unauthenticated to allow bootstrapping
// a fresh installation. In production deployments it should be protected by a
// network-level control or an admin secret header before being exposed publicly.
func (h *TenantsHandler) CreateTenant(w http.ResponseWriter, r *http.Request) {
	var req createTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" || req.Slug == "" {
		writeError(w, http.StatusBadRequest, "name and slug are required")
		return
	}

	tenant, err := h.store.CreateTenant(r.Context(), req.Name, req.Slug)
	if err != nil {
		h.log.Error("create tenant failed", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to create tenant")
		return
	}

	writeJSON(w, http.StatusCreated, tenant)
}

// createAPIKeyRequest is the expected POST body for key creation.
type createAPIKeyRequest struct {
	Name   string   `json:"name"`
	Scopes []string `json:"scopes"`
}

// CreateAPIKey handles POST /v1/tenants/{tenantSlug}/api-keys.
// Requires "admin" scope on the calling key.
func (h *TenantsHandler) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	tenantSlug := chi.URLParam(r, "tenantSlug")

	tenant, err := h.store.GetTenantBySlug(r.Context(), tenantSlug)
	if err != nil {
		writeError(w, http.StatusNotFound, "tenant not found")
		return
	}

	var req createAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if len(req.Scopes) == 0 {
		req.Scopes = []string{"invoke"}
	}

	key, plain, err := h.store.CreateAPIKeyWithPlain(r.Context(), tenant.ID, req.Name, req.Scopes)
	if err != nil {
		h.log.Error("create api key failed", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to create api key")
		return
	}

	// Return the plain key only once — it cannot be recovered after this.
	writeJSON(w, http.StatusCreated, map[string]any{
		"id":          key.ID,
		"tenant_id":   key.TenantID,
		"name":        key.Name,
		"scopes":      key.Scopes,
		"key":         plain, // one-time reveal
		"key_prefix":  key.KeyPrefix,
		"created_at":  key.CreatedAt,
	})
}

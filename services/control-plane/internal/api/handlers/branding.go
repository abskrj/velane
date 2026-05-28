package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/runeforge/control-plane/internal/models"
	"go.uber.org/zap"
)

// BrandingStore is the subset of *postgres.Store that branding handlers need.
type BrandingStore interface {
	GetTenantBySlug(ctx context.Context, slug string) (*models.Tenant, error)
	GetBranding(ctx context.Context, tenantID string) (*models.Branding, error)
	UpdateBranding(ctx context.Context, tenantID string, b models.Branding) error
}

// BrandingHandler handles branding config endpoints.
type BrandingHandler struct {
	store BrandingStore
	log   *zap.Logger
}

// NewBrandingHandler constructs a BrandingHandler.
func NewBrandingHandler(store BrandingStore, log *zap.Logger) *BrandingHandler {
	return &BrandingHandler{store: store, log: log}
}

// GetBranding handles GET /v1/tenants/{slug}/branding.
func (h *BrandingHandler) GetBranding(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "tenantSlug")
	tenant, err := h.store.GetTenantBySlug(r.Context(), slug)
	if err != nil {
		writeError(w, http.StatusNotFound, "tenant not found")
		return
	}

	branding, err := h.store.GetBranding(r.Context(), tenant.ID)
	if err != nil {
		h.log.Error("get branding failed", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to get branding")
		return
	}

	writeJSON(w, http.StatusOK, branding)
}

// UpdateBranding handles PUT /v1/tenants/{slug}/branding.
func (h *BrandingHandler) UpdateBranding(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "tenantSlug")
	tenant, err := h.store.GetTenantBySlug(r.Context(), slug)
	if err != nil {
		writeError(w, http.StatusNotFound, "tenant not found")
		return
	}

	var b models.Branding
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.store.UpdateBranding(r.Context(), tenant.ID, b); err != nil {
		h.log.Error("update branding failed", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to update branding")
		return
	}

	writeJSON(w, http.StatusOK, b)
}

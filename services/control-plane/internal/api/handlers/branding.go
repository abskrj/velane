package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/abskrj/velane/services/control-plane/internal/api/middleware"
	"github.com/abskrj/velane/services/control-plane/internal/audit"
	"github.com/abskrj/velane/services/control-plane/internal/models"
	"go.uber.org/zap"
)

// BrandingStore is the subset of *postgres.Store that branding handlers need.
type BrandingStore interface {
	GetBranding(ctx context.Context, tenantID string) (*models.Branding, error)
	UpdateBranding(ctx context.Context, tenantID string, b models.Branding) error
}

// BrandingHandler handles branding config endpoints.
type BrandingHandler struct {
	store   BrandingStore
	log     *zap.Logger
	auditor *audit.Logger
}

// NewBrandingHandler constructs a BrandingHandler.
func NewBrandingHandler(store BrandingStore, log *zap.Logger) *BrandingHandler {
	return &BrandingHandler{store: store, log: log}
}

// WithAuditor attaches an audit logger to the BrandingHandler.
func (h *BrandingHandler) WithAuditor(a *audit.Logger) *BrandingHandler {
	h.auditor = a
	return h
}

// GetBranding handles GET /v1/tenant/branding.
func (h *BrandingHandler) GetBranding(w http.ResponseWriter, r *http.Request) {
	authTenant := middleware.TenantFromContext(r.Context())
	if authTenant == nil {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	branding, err := h.store.GetBranding(r.Context(), authTenant.ID)
	if err != nil {
		h.log.Error("get branding failed", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to get branding")
		return
	}

	writeJSON(w, http.StatusOK, branding)
}

// UpdateBranding handles PUT /v1/tenant/branding.
func (h *BrandingHandler) UpdateBranding(w http.ResponseWriter, r *http.Request) {
	authTenant := middleware.TenantFromContext(r.Context())
	if authTenant == nil {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	var b models.Branding
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.store.UpdateBranding(r.Context(), authTenant.ID, b); err != nil {
		h.log.Error("update branding failed", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to update branding")
		return
	}

	if h.auditor != nil {
		actorID, actorType := resolveActor(r)
		h.auditor.Log(r.Context(), models.AuditEntry{
			TenantID:   authTenant.ID,
			ActorID:    actorID,
			ActorType:  actorType,
			Action:     "branding_update",
			ResourceID: authTenant.ID,
		})
	}

	writeJSON(w, http.StatusOK, b)
}

package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/abskrj/velane/services/control-plane/internal/api/middleware"
	"github.com/abskrj/velane/services/control-plane/internal/models"
	"github.com/abskrj/velane/services/control-plane/internal/store/postgres"
	"go.uber.org/zap"
)

// AuditStore is the subset of *postgres.Store that audit handlers need.
type AuditStore interface {
	ListAuditLog(ctx context.Context, tenantID string, opts postgres.AuditQueryOpts) ([]*models.AuditEntry, error)
}

// AuditHandler handles audit log endpoints.
type AuditHandler struct {
	store AuditStore
	log   *zap.Logger
}

// NewAuditHandler constructs an AuditHandler.
func NewAuditHandler(store AuditStore, log *zap.Logger) *AuditHandler {
	return &AuditHandler{store: store, log: log}
}

// ListAuditLog handles GET /v1/tenant/audit-log.
// Query params: limit (default 50, max 200), action (optional), before (RFC3339 cursor).
// Requires admin scope.
func (h *AuditHandler) ListAuditLog(w http.ResponseWriter, r *http.Request) {
	authTenant := middleware.TenantFromContext(r.Context())
	if authTenant == nil {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	opts := postgres.AuditQueryOpts{}

	if v := r.URL.Query().Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err == nil && n > 0 {
			opts.Limit = n
		}
	}
	if opts.Limit == 0 {
		opts.Limit = 50
	}

	if v := r.URL.Query().Get("action"); v != "" {
		opts.Action = v
	}

	if v := r.URL.Query().Get("before"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err == nil {
			opts.Before = t
		}
	}

	entries, err := h.store.ListAuditLog(r.Context(), authTenant.ID, opts)
	if err != nil {
		h.log.Error("list audit log failed", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to list audit log")
		return
	}

	if entries == nil {
		writeJSON(w, http.StatusOK, []*models.AuditEntry{})
		return
	}
	writeJSON(w, http.StatusOK, entries)
}

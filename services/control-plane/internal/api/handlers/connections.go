package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/abskrj/velane/services/control-plane/internal/api/middleware"
	"github.com/abskrj/velane/services/control-plane/internal/audit"
	"github.com/abskrj/velane/services/control-plane/internal/models"
	"github.com/abskrj/velane/services/control-plane/internal/nango"
	"github.com/abskrj/velane/services/control-plane/internal/store/postgres"
	"go.uber.org/zap"
)

// ConnectionStore is the subset of postgres.Store needed by ConnectionsHandler.
type ConnectionStore interface {
	GetTenantBySlug(ctx context.Context, slug string) (*models.Tenant, error)
	UpsertConnection(ctx context.Context, tenantID, provider, alias, displayName string) (*models.Connection, error)
	ListConnections(ctx context.Context, tenantID string) ([]*models.Connection, error)
	GetConnection(ctx context.Context, tenantID, provider string) (*models.Connection, error)
	DeleteConnection(ctx context.Context, tenantID, provider string) error
}

// ConnectionsHandler handles all OAuth connection management endpoints
// and the internal proxy that snippet code calls at runtime.
type ConnectionsHandler struct {
	store           *postgres.Store
	nango           *nango.Client
	log             *zap.Logger
	auditor         *audit.Logger
	nangoConnectURL string // browser-accessible Connect UI URL, returned with session tokens
	nangoApiURL     string // browser-accessible Nango API URL, returned with session tokens
}

func NewConnectionsHandler(store *postgres.Store, nangoClient *nango.Client, log *zap.Logger, nangoConnectURL, nangoApiURL string) *ConnectionsHandler {
	return &ConnectionsHandler{store: store, nango: nangoClient, log: log, nangoConnectURL: nangoConnectURL, nangoApiURL: nangoApiURL}
}

func (h *ConnectionsHandler) WithAuditor(a *audit.Logger) *ConnectionsHandler {
	h.auditor = a
	return h
}

// CreateSession handles POST /v1/tenants/{slug}/connections/session.
// Returns a short-lived Nango Connect session token the frontend uses to open
// the OAuth popup for a specific provider.
func (h *ConnectionsHandler) CreateSession(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.TenantFromContext(r.Context())
	if tenant == nil {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	authTenant, err := h.store.GetTenantBySlug(r.Context(), chi.URLParam(r, "tenantSlug"))
	if err != nil {
		writeError(w, http.StatusNotFound, "tenant not found")
		return
	}
	if authTenant.ID != tenant.ID {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	var req struct {
		Provider string `json:"provider"`
		Alias    string `json:"alias"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Provider == "" {
		writeError(w, http.StatusBadRequest, "provider is required")
		return
	}
	if req.Alias == "" {
		req.Alias = "default"
	}

	token, err := h.nango.CreateConnectSession(r.Context(), tenant.ID, tenant.Name, req.Provider, req.Alias)
	if err != nil {
		h.log.Error("create nango connect session", zap.String("provider", req.Provider), zap.Error(err))
		writeError(w, http.StatusBadGateway, "failed to create connect session")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"session_token": token,
		"connect_url":   h.nangoConnectURL,
		"api_url":       h.nangoApiURL,
	})
}

// RecordConnection handles POST /v1/tenants/{slug}/connections.
// Called by the frontend after the Nango OAuth popup completes successfully.
func (h *ConnectionsHandler) RecordConnection(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.TenantFromContext(r.Context())
	if tenant == nil {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	authTenant, err := h.store.GetTenantBySlug(r.Context(), chi.URLParam(r, "tenantSlug"))
	if err != nil {
		writeError(w, http.StatusNotFound, "tenant not found")
		return
	}
	if authTenant.ID != tenant.ID {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	var req struct {
		Provider    string `json:"provider"`
		DisplayName string `json:"display_name"`
		Alias       string `json:"alias"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Provider == "" {
		writeError(w, http.StatusBadRequest, "provider is required")
		return
	}
	if req.Alias == "" {
		req.Alias = "default"
	}

	conn, err := h.store.UpsertConnection(r.Context(), tenant.ID, req.Provider, req.Alias, req.DisplayName)
	if err != nil {
		h.log.Error("record connection", zap.String("provider", req.Provider), zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to record connection")
		return
	}

	if h.auditor != nil {
		actorID, actorType := resolveActor(r)
		h.auditor.Log(r.Context(), models.AuditEntry{
			TenantID:   tenant.ID,
			ActorID:    actorID,
			ActorType:  actorType,
			Action:     "connection_connect",
			ResourceID: conn.ID,
			Metadata:   auditMeta(map[string]any{"provider": req.Provider}),
		})
	}

	writeJSON(w, http.StatusCreated, conn)
}

// ListConnections handles GET /v1/tenants/{slug}/connections.
func (h *ConnectionsHandler) ListConnections(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.TenantFromContext(r.Context())
	if tenant == nil {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	authTenant, err := h.store.GetTenantBySlug(r.Context(), chi.URLParam(r, "tenantSlug"))
	if err != nil {
		writeError(w, http.StatusNotFound, "tenant not found")
		return
	}
	if authTenant.ID != tenant.ID {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	conns, err := h.store.ListConnections(r.Context(), tenant.ID)
	if err != nil {
		h.log.Error("list connections", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to list connections")
		return
	}

	if conns == nil {
		writeJSON(w, http.StatusOK, []any{})
		return
	}
	writeJSON(w, http.StatusOK, conns)
}

// DisconnectProvider handles DELETE /v1/tenants/{slug}/connections/{provider}.
func (h *ConnectionsHandler) DisconnectProvider(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.TenantFromContext(r.Context())
	if tenant == nil {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	authTenant, err := h.store.GetTenantBySlug(r.Context(), chi.URLParam(r, "tenantSlug"))
	if err != nil {
		writeError(w, http.StatusNotFound, "tenant not found")
		return
	}
	if authTenant.ID != tenant.ID {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	provider := chi.URLParam(r, "provider")

	conn, err := h.store.GetConnection(r.Context(), tenant.ID, provider)
	if err != nil {
		writeError(w, http.StatusNotFound, "connection not found")
		return
	}

	// Delete from Nango first (best-effort — don't block if Nango is down).
	if err := h.nango.DeleteConnection(r.Context(), conn.NangoConnectionID, provider); err != nil {
		h.log.Warn("nango delete connection failed (continuing)", zap.String("provider", provider), zap.Error(err))
	}

	if err := h.store.DeleteConnection(r.Context(), tenant.ID, provider); err != nil {
		h.log.Error("delete connection", zap.String("provider", provider), zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to delete connection")
		return
	}

	if h.auditor != nil {
		actorID, actorType := resolveActor(r)
		h.auditor.Log(r.Context(), models.AuditEntry{
			TenantID:   tenant.ID,
			ActorID:    actorID,
			ActorType:  actorType,
			Action:     "connection_disconnect",
			ResourceID: conn.ID,
			Metadata:   auditMeta(map[string]any{"provider": provider}),
		})
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListConnectionsForToken handles GET /v1/connections (no tenant slug in path).
// Used by the MCP server, which authenticates with an API key and does not
// need to know the tenant slug in advance.
func (h *ConnectionsHandler) ListConnectionsForToken(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.TenantFromContext(r.Context())
	if tenant == nil {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	conns, err := h.store.ListConnections(r.Context(), tenant.ID)
	if err != nil {
		h.log.Error("list connections for token", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to list connections")
		return
	}

	if conns == nil {
		writeJSON(w, http.StatusOK, []any{})
		return
	}
	writeJSON(w, http.StatusOK, conns)
}

// Proxy handles all methods on /v1/proxy/{provider}/*.
// This endpoint is intentionally unauthenticated via the public middleware stack —
// it is only reachable from the internal Docker network (executor containers).
// It trusts the X-Velane-Tenant header, which is set by the executor runtime from
// the VELANE_TENANT_ID env var injected at invocation time.
func (h *ConnectionsHandler) Proxy(w http.ResponseWriter, r *http.Request) {
	tenantID := strings.TrimSpace(r.Header.Get("X-Velane-Tenant"))
	if tenantID == "" {
		writeError(w, http.StatusBadRequest, "X-Velane-Tenant header required")
		return
	}

	provider := chi.URLParam(r, "provider")
	// chi wildcard gives us the path after {provider}/ — re-add leading slash.
	path := "/" + chi.URLParam(r, "*")

	conn, err := h.store.GetConnection(r.Context(), tenantID, provider)
	if err != nil {
		writeError(w, http.StatusBadRequest, "no connection found for provider: "+provider)
		return
	}

	h.nango.Proxy(w, r, conn.NangoConnectionID, provider, path)
}

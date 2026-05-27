package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/runeforge/control-plane/internal/api/middleware"
	"github.com/runeforge/control-plane/internal/scheduler"
	"github.com/runeforge/control-plane/internal/store/postgres"
	"go.uber.org/zap"
)

// InvocationsHandler bundles invocation-related HTTP handlers.
type InvocationsHandler struct {
	store     *postgres.Store
	scheduler *scheduler.Scheduler
	log       *zap.Logger
}

// NewInvocationsHandler constructs an InvocationsHandler.
func NewInvocationsHandler(store *postgres.Store, sched *scheduler.Scheduler, log *zap.Logger) *InvocationsHandler {
	return &InvocationsHandler{store: store, scheduler: sched, log: log}
}

// Invoke handles POST /v1/invoke/{tenantSlug}/{snippetSlug}.
//
// The invoke endpoint performs its own API key authentication inline rather
// than relying on the global auth middleware. This is because the tenant is
// identified via the URL path, allowing external callers to use a key scoped
// to a specific tenant without needing a separate header.
//
// Response headers:
//
//	X-Invocation-Id  — the persisted invocation ID
//	X-Duration-Ms    — execution wall-clock time in milliseconds
func (h *InvocationsHandler) Invoke(w http.ResponseWriter, r *http.Request) {
	tenantSlug := chi.URLParam(r, "tenantSlug")
	snippetSlug := chi.URLParam(r, "snippetSlug")

	// --- Inline API key auth ---
	authHdr := r.Header.Get("Authorization")
	if authHdr == "" {
		writeError(w, http.StatusUnauthorized, "missing Authorization header")
		return
	}
	parts := strings.SplitN(authHdr, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		writeError(w, http.StatusUnauthorized, "malformed Authorization header")
		return
	}
	plainKey := strings.TrimSpace(parts[1])

	key, err := h.store.ValidateAPIKey(r.Context(), plainKey)
	if err != nil {
		h.log.Debug("invoke: invalid api key", zap.Error(err))
		writeError(w, http.StatusUnauthorized, "invalid api key")
		return
	}

	if !key.HasScope("invoke") {
		writeError(w, http.StatusForbidden, "api key missing 'invoke' scope")
		return
	}

	// Resolve the tenant from the URL and verify the key belongs to it.
	tenant, err := h.store.GetTenantBySlug(r.Context(), tenantSlug)
	if err != nil {
		writeError(w, http.StatusNotFound, "tenant not found")
		return
	}
	if key.TenantID != tenant.ID {
		writeError(w, http.StatusForbidden, "api key does not belong to this tenant")
		return
	}

	// --- Read the input payload ---
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20)) // 1 MiB limit
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read request body")
		return
	}
	inputPayload := string(body)
	if inputPayload == "" {
		inputPayload = "{}"
	}

	// Validate that the body is valid JSON (we store it raw).
	if !json.Valid([]byte(inputPayload)) {
		writeError(w, http.StatusBadRequest, "request body must be valid JSON")
		return
	}

	// --- Resolve env from query param, default to prod ---
	env := r.URL.Query().Get("env")
	if env == "" {
		env = "prod"
	}
	if env != "dev" && env != "prod" {
		writeError(w, http.StatusBadRequest, "env must be 'dev' or 'prod'")
		return
	}

	// --- Execute via scheduler ---
	invocation, err := h.scheduler.Invoke(r.Context(), scheduler.InvokeRequest{
		TenantID:    tenant.ID,
		SnippetSlug: snippetSlug,
		Env:         env,
		Input:       inputPayload,
	})
	if err != nil {
		h.log.Error("invoke failed", zap.String("slug", snippetSlug), zap.Error(err))
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// --- Set response headers ---
	w.Header().Set("X-Invocation-Id", invocation.ID)
	w.Header().Set("X-Duration-Ms", fmt.Sprintf("%d", invocation.DurationMs))

	// Parse the output JSON so it embeds naturally in the response object.
	// If it's not valid JSON, embed it as a string.
	var outputVal any
	if err := json.Unmarshal([]byte(invocation.Output), &outputVal); err != nil {
		outputVal = invocation.Output
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"output":        outputVal,
		"invocation_id": invocation.ID,
		"duration_ms":   invocation.DurationMs,
		"status":        invocation.Status,
		"error":         invocation.Error,
		"stderr":        invocation.Stderr,
	})
}

// GetInvocation handles GET /v1/invocations/{id}.
func (h *InvocationsHandler) GetInvocation(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.TenantFromContext(r.Context())
	if tenant == nil {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	id := chi.URLParam(r, "id")
	invocation, err := h.store.GetInvocation(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "invocation not found")
		return
	}

	// Enforce tenant isolation.
	if invocation.TenantID != tenant.ID {
		writeError(w, http.StatusNotFound, "invocation not found")
		return
	}

	writeJSON(w, http.StatusOK, invocation)
}

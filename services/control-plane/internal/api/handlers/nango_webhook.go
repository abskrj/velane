package handlers

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"

	"github.com/abskrj/velane/services/control-plane/internal/models"
	"github.com/abskrj/velane/services/control-plane/internal/nango"
	"go.uber.org/zap"
)

// NangoWebhookStore is the store subset needed by NangoWebhookHandler.
type NangoWebhookStore interface {
	UpdateNangoConnectionID(ctx context.Context, tenantID, provider, alias, nangoConnID string) (*models.Connection, error)
	UpdateNangoConnectionIDByProviderConfigKey(ctx context.Context, tenantID, providerConfigKey, nangoConnID string) (*models.Connection, error)
}

// NangoWebhookHandler handles POST /v1/webhooks/nango.
type NangoWebhookHandler struct {
	store         NangoWebhookStore
	nango         *nango.Client
	webhookSecret string
	log           *zap.Logger
}

func NewNangoWebhookHandler(store NangoWebhookStore, nangoClient *nango.Client, webhookSecret string, log *zap.Logger) *NangoWebhookHandler {
	return &NangoWebhookHandler{store: store, nango: nangoClient, webhookSecret: webhookSecret, log: log}
}

// nangoWebhookPayload is the subset of Nango's webhook body that we care about.
//
// ⚠️ Field names must be verified against the actual Nango self-hosted webhook delivery.
// Fire a real OAuth flow and inspect the raw body in container logs to confirm:
//   - Top-level field casing (camelCase shown here is what Nango JS SDK uses)
//   - The path where the connection ID is returned ("connectionId" assumed)
//   - The signature header name (X-Nango-Signature assumed; raw hex, no prefix)
type nangoWebhookPayload struct {
	Type              string `json:"type"`              // "auth"
	ConnectionID      string `json:"connectionId"`      // Nango-generated UUID
	ProviderConfigKey string `json:"providerConfigKey"` // integration key e.g. "salesforce"
	Success           bool   `json:"success"`
	EndUser           struct {
		ID          string `json:"id"` // our tenant_id, set in CreateConnectSession
		DisplayName string `json:"displayName"`
	} `json:"endUser"`
}

// HandleNangoEvent receives Nango webhook events and stores the real connection UUID.
func (h *NangoWebhookHandler) HandleNangoEvent(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read body")
		return
	}

	if h.webhookSecret != "" {
		sig := r.Header.Get("X-Nango-Signature")
		if !verifyNangoSignature(body, h.webhookSecret, sig) {
			writeError(w, http.StatusUnauthorized, "invalid webhook signature")
			return
		}
	} else {
		h.log.Warn("NANGO_WEBHOOK_SECRET not set — skipping signature verification")
	}

	var payload nangoWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	// Only handle successful auth events; acknowledge everything else silently.
	if payload.Type != "auth" || !payload.Success {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ignored"})
		return
	}

	tenantID := payload.EndUser.ID
	provider := payload.ProviderConfigKey
	nangoConnID := payload.ConnectionID

	if tenantID == "" || provider == "" || nangoConnID == "" {
		writeError(w, http.StatusBadRequest, "missing required fields in webhook payload")
		return
	}

	conn, err := h.store.UpdateNangoConnectionIDByProviderConfigKey(r.Context(), tenantID, provider, nangoConnID)
	if err != nil {
		// The row may not exist yet if the webhook beats the frontend's RecordConnection call
		// (a rare race). Return 200 so Nango does not retry — the frontend will create the row
		// shortly after and the user can re-authenticate to get the UUID written.
		h.log.Warn("nango webhook: connection row not found — RecordConnection not yet called",
			zap.String("tenant_id", tenantID),
			zap.String("provider", provider),
			zap.Error(err),
		)
		writeJSON(w, http.StatusOK, map[string]string{"status": "deferred"})
		return
	}

	// Write the alias back to Nango as connection metadata so it's visible on the
	// Nango side. Best-effort — a failure here does not affect our own DB state.
	if h.nango != nil {
		providerConfigKey := conn.ProviderConfigKey
		if providerConfigKey == "" {
			providerConfigKey = provider
		}
		if err := h.nango.PatchConnectionMetadata(r.Context(), nangoConnID, providerConfigKey, map[string]any{
			"velane_alias":     conn.Alias,
			"velane_tenant_id": tenantID,
		}); err != nil {
			h.log.Warn("nango webhook: failed to patch connection metadata (non-fatal)",
				zap.String("nango_connection_id", nangoConnID),
				zap.Error(err),
			)
		}
	}

	h.log.Info("nango webhook: connection ID stored",
		zap.String("tenant_id", tenantID),
		zap.String("provider", provider),
		zap.String("alias", conn.Alias),
		zap.String("nango_connection_id", nangoConnID),
	)
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// verifyNangoSignature checks the HMAC-SHA256 signature Nango attaches to webhook deliveries.
// Nango self-hosted sends a raw hex digest (no "sha256=" prefix) in X-Nango-Signature.
// ⚠️ Verify the exact format against your Nango version before enabling in production.
func verifyNangoSignature(body []byte, secret, sig string) bool {
	if sig == "" {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(sig))
}

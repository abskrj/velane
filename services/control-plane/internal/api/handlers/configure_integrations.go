package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/abskrj/velane/services/control-plane/internal/nango"
	"go.uber.org/zap"
)

// ConfigureIntegrationsHandler manages operator-level integration configs
// (OAuth app credentials stored in Nango). These are platform-wide, not per-tenant.
type ConfigureIntegrationsHandler struct {
	nango *nango.Client
	log   *zap.Logger
}

func NewConfigureIntegrationsHandler(nangoClient *nango.Client, log *zap.Logger) *ConfigureIntegrationsHandler {
	return &ConfigureIntegrationsHandler{nango: nangoClient, log: log}
}

// ListConfigured handles GET /v1/integrations/configured.
// Returns all provider configs (OAuth apps) set up in Nango.
func (h *ConfigureIntegrationsHandler) ListConfigured(w http.ResponseWriter, r *http.Request) {
	configs, err := h.nango.ListIntegrationConfigs(r.Context())
	if err != nil {
		h.log.Error("list integration configs", zap.Error(err))
		writeError(w, http.StatusBadGateway, "failed to list configured integrations")
		return
	}
	if configs == nil {
		writeJSON(w, http.StatusOK, []any{})
		return
	}
	writeJSON(w, http.StatusOK, configs)
}

// Configure handles POST /v1/integrations/configured.
// Creates or updates a provider config in Nango.
// For OAUTH2: requires oauth_client_id and oauth_client_secret.
// For API_KEY/BASIC: provider and provider_config_key are sufficient.
func (h *ConfigureIntegrationsHandler) Configure(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ProviderConfigKey string `json:"provider_config_key"`
		Provider          string `json:"provider"`
		OAuthClientID     string `json:"oauth_client_id"`
		OAuthClientSecret string `json:"oauth_client_secret"`
		OAuthScopes       string `json:"oauth_scopes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Provider == "" {
		writeError(w, http.StatusBadRequest, "provider is required")
		return
	}
	// Default provider_config_key to provider slug.
	if req.ProviderConfigKey == "" {
		req.ProviderConfigKey = req.Provider
	}

	if err := h.nango.CreateIntegrationConfig(
		r.Context(),
		req.ProviderConfigKey,
		req.Provider,
		req.OAuthClientID,
		req.OAuthClientSecret,
		req.OAuthScopes,
	); err != nil {
		h.log.Error("configure integration", zap.String("provider", req.Provider), zap.Error(err))
		writeError(w, http.StatusBadGateway, "failed to configure integration: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"provider_config_key": req.ProviderConfigKey,
		"provider":            req.Provider,
	})
}

// DeleteConfigured handles DELETE /v1/integrations/configured/{providerConfigKey}.
func (h *ConfigureIntegrationsHandler) DeleteConfigured(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "providerConfigKey")
	if err := h.nango.DeleteIntegrationConfig(r.Context(), key); err != nil {
		h.log.Error("delete integration config", zap.String("key", key), zap.Error(err))
		writeError(w, http.StatusBadGateway, "failed to delete integration config")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

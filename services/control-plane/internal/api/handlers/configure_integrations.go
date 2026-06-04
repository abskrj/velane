package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/abskrj/velane/services/control-plane/internal/api/middleware"
	"github.com/abskrj/velane/services/control-plane/internal/nango"
	"github.com/abskrj/velane/services/control-plane/internal/store/postgres"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type ConfigureIntegrationsHandler struct {
	nango  *nango.Client
	log    *zap.Logger
	store  *postgres.Store
	encKey []byte
}

func NewConfigureIntegrationsHandler(store *postgres.Store, nangoClient *nango.Client, log *zap.Logger, encKey []byte) *ConfigureIntegrationsHandler {
	return &ConfigureIntegrationsHandler{store: store, nango: nangoClient, log: log, encKey: encKey}
}

// ListConfigured handles GET /v1/integrations/configured.
// Returns all integration credential profiles for the authenticated tenant.
func (h *ConfigureIntegrationsHandler) ListConfigured(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.TenantFromContext(r.Context())
	if tenant == nil {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	configs, err := h.store.ListIntegrationCredentialProfiles(r.Context(), tenant.ID, "")
	if err != nil {
		h.log.Error("list integration credential profiles", zap.Error(err))
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
func (h *ConfigureIntegrationsHandler) Configure(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.TenantFromContext(r.Context())
	if tenant == nil {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	var req struct {
		Provider          string            `json:"provider"`
		Alias             string            `json:"alias"`
		Name              string            `json:"name"`
		CredentialsType   string            `json:"credentials_type"`
		Credentials       map[string]string `json:"credentials"`
		OAuthClientID     string            `json:"oauth_client_id"`
		OAuthClientSecret string            `json:"oauth_client_secret"`
		OAuthScopes       string            `json:"oauth_scopes"`
		IsDefault         bool              `json:"is_default"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Provider == "" {
		writeError(w, http.StatusBadRequest, "provider is required")
		return
	}
	if req.Alias == "" {
		req.Alias = "default"
	}
	if req.Name == "" {
		req.Name = req.Alias
	}

	credType := strings.ToUpper(strings.TrimSpace(req.CredentialsType))
	if np, err := h.nango.GetProvider(r.Context(), req.Provider); err == nil {
		if providerMode := strings.ToUpper(strings.TrimSpace(np.AuthMode)); providerMode != "" {
			if credType != "" && credType != providerMode {
				h.log.Warn("credentials_type overridden by provider auth_mode",
					zap.String("provider", req.Provider),
					zap.String("requested", credType),
					zap.String("resolved", providerMode),
				)
			}
			credType = providerMode
		}
	} else {
		h.log.Debug("nango provider lookup failed; using request credentials_type",
			zap.String("provider", req.Provider),
			zap.Error(err),
		)
	}
	if credType == "" {
		credType = "OAUTH2"
	}

	plainFields := map[string]string{}
	for k, v := range req.Credentials {
		if strings.TrimSpace(k) == "" || strings.TrimSpace(v) == "" {
			continue
		}
		plainFields[k] = v
	}
	if req.OAuthClientID != "" {
		plainFields["oauth_client_id"] = req.OAuthClientID
	}
	if req.OAuthClientSecret != "" {
		plainFields["oauth_client_secret"] = req.OAuthClientSecret
	}
	if req.OAuthScopes != "" {
		plainFields["oauth_scopes"] = req.OAuthScopes
	}
	resolvedOAuthScopes := firstNonEmpty(plainFields["oauth_scopes"], plainFields["scopes"])

	existing, _ := h.store.GetIntegrationCredentialProfileByAlias(r.Context(), tenant.ID, req.Provider, req.Alias)
	providerConfigKey := ""
	if existing != nil {
		providerConfigKey = existing.NangoProviderConfigKey
	}

	nangoCreds := map[string]any{"type": credType}
	switch credType {
	case "OAUTH2":
		clientID := firstNonEmpty(plainFields["oauth_client_id"], plainFields["client_id"])
		clientSecret := firstNonEmpty(plainFields["oauth_client_secret"], plainFields["client_secret"])
		if clientID == "" || clientSecret == "" {
			writeError(w, http.StatusBadRequest, "client_id and client_secret are required")
			return
		}
		nangoCreds["client_id"] = clientID
		nangoCreds["client_secret"] = clientSecret
		if scopes := firstNonEmpty(plainFields["oauth_scopes"], plainFields["scopes"]); scopes != "" {
			nangoCreds["scopes"] = parseScopes(scopes)
		}
	case "OAUTH2_CC":
		clientID := firstNonEmpty(plainFields["oauth_client_id"], plainFields["client_id"])
		clientSecret := firstNonEmpty(plainFields["oauth_client_secret"], plainFields["client_secret"])
		if clientID == "" || clientSecret == "" {
			writeError(w, http.StatusBadRequest, "client_id and client_secret are required")
			return
		}
		// Nango expects OAUTH2_CC providers to be created without credentials in
		// /integrations; tenant credentials are still securely stored in Velane.
		nangoCreds = map[string]any{}
	default:
		for k, v := range plainFields {
			nangoCreds[k] = v
		}
	}

	if err := h.nango.CreateIntegrationConfig(
		r.Context(),
		chooseProviderConfigKey(providerConfigKey, tenant.ID, req.Provider, req.Alias),
		req.Provider,
		nangoCreds,
	); err != nil {
		h.log.Error("configure integration", zap.String("provider", req.Provider), zap.Error(err))
		writeError(w, http.StatusBadGateway, "failed to configure integration: "+err.Error())
		return
	}

	profile, err := h.store.UpsertIntegrationCredentialProfile(
		r.Context(),
		tenant.ID,
		req.Provider,
		req.Alias,
		req.Name,
		resolveActorID(r),
		credType,
		plainFields,
		resolvedOAuthScopes,
		req.IsDefault,
		chooseProviderConfigKey(providerConfigKey, tenant.ID, req.Provider, req.Alias),
		h.encKey,
	)
	if err != nil {
		h.log.Error("upsert integration credential profile", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to store integration credentials")
		return
	}

	view, _, err := h.store.DecryptIntegrationCredentialProfile(profile, h.encKey)
	if err != nil {
		h.log.Error("decrypt integration profile view", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to return integration profile")
		return
	}
	writeJSON(w, http.StatusOK, view)
}

// DeleteConfigured handles DELETE /v1/integrations/configured/{id}.
func (h *ConfigureIntegrationsHandler) DeleteConfigured(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.TenantFromContext(r.Context())
	if tenant == nil {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	id := chi.URLParam(r, "providerConfigKey")
	profile, err := h.store.SoftDeleteIntegrationCredentialProfile(r.Context(), tenant.ID, id)
	if err != nil {
		h.log.Error("soft delete integration credential profile", zap.String("id", id), zap.Error(err))
		writeError(w, http.StatusNotFound, "integration credential profile not found")
		return
	}

	if err := h.nango.DeleteIntegrationConfig(r.Context(), profile.NangoProviderConfigKey); err != nil {
		h.log.Error("delete integration config", zap.String("key", profile.NangoProviderConfigKey), zap.Error(err))
		writeError(w, http.StatusBadGateway, "failed to delete integration config")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func chooseProviderConfigKey(existing, tenantID, provider, alias string) string {
	if strings.TrimSpace(existing) != "" {
		return existing
	}
	return fmt.Sprintf("velane_%s_%s_%s", sanitizeConfigPart(tenantID), sanitizeConfigPart(provider), sanitizeConfigPart(alias))
}

func sanitizeConfigPart(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	v = strings.ReplaceAll(v, " ", "_")
	v = strings.ReplaceAll(v, "-", "_")
	if v == "" {
		return "default"
	}
	return v
}

func resolveActorID(r *http.Request) string {
	if user := middleware.SessionUserFromContext(r.Context()); user != nil {
		return user.ID
	}
	if key := middleware.APIKeyFromContext(r.Context()); key != nil {
		return key.ID
	}
	return ""
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if trimmed := strings.TrimSpace(v); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func parseScopes(raw string) []string {
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t' || r == '\n'
	})
	out := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, part := range parts {
		scope := strings.TrimSpace(part)
		if scope == "" {
			continue
		}
		if _, ok := seen[scope]; ok {
			continue
		}
		seen[scope] = struct{}{}
		out = append(out, scope)
	}
	return out
}

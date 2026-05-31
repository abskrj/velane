package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/abskrj/velane/services/control-plane/internal/models"
	"github.com/abskrj/velane/services/control-plane/internal/nango"
	"github.com/abskrj/velane/services/control-plane/internal/providers"
	"go.uber.org/zap"
)

// IntegrationsHandler serves the provider catalog and per-provider documentation.
type IntegrationsHandler struct {
	nango            *nango.Client
	log              *zap.Logger
	nangoInternalURL string // used to proxy Nango assets through control plane
}

func NewIntegrationsHandler(nangoClient *nango.Client, log *zap.Logger, nangoInternalURL string) *IntegrationsHandler {
	return &IntegrationsHandler{nango: nangoClient, log: log, nangoInternalURL: nangoInternalURL}
}

// rewriteLogoURL replaces Nango's self-reported logo URL (which points back at
// Nango directly) with a path served through the control plane proxy so the
// browser never needs to reach Nango.
func (h *IntegrationsHandler) rewriteLogoURL(nangoLogoURL string) string {
	if nangoLogoURL == "" {
		return ""
	}
	// Extract just the path portion (e.g. /images/template-logos/github.svg)
	// regardless of what host Nango reports.
	for _, prefix := range []string{"http://", "https://"} {
		if strings.HasPrefix(nangoLogoURL, prefix) {
			rest := strings.TrimPrefix(nangoLogoURL, prefix)
			idx := strings.Index(rest, "/")
			if idx >= 0 {
				return "/v1/nango-assets" + rest[idx:]
			}
		}
	}
	return nangoLogoURL
}

// ProxyAsset handles GET /v1/nango-assets/* and proxies static assets
// (logos, icons) from Nango's internal server to the browser.
// No auth required — these are public image files.
func (h *IntegrationsHandler) ProxyAsset(w http.ResponseWriter, r *http.Request) {
	assetPath := chi.URLParam(r, "*")
	url := h.nangoInternalURL + "/" + assetPath

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, url, nil)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "upstream error", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if ct := resp.Header.Get("Content-Type"); ct != "" {
		w.Header().Set("Content-Type", ct)
	}
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body) //nolint:errcheck
}

// ListProviders handles GET /v1/integrations.
// Returns all available provider templates from Nango (cached 1h).
// Falls back to the bundled catalog if Nango is unavailable.
// Always returns the same shape: { unique_key, name, auth_mode, categories, logo_url }.
func (h *IntegrationsHandler) ListProviders(w http.ResponseWriter, r *http.Request) {
	list, err := h.nango.ListProviders(r.Context())
	if err != nil {
		h.log.Warn("nango /providers unavailable, falling back to bundled catalog", zap.Error(err))
		writeJSON(w, http.StatusOK, providers.Catalog)
		return
	}

	// Normalise to a stable API shape. Include connection_config and credentials
	// so the frontend can build dynamic configuration forms.
	type providerOut struct {
		UniqueKey        string                            `json:"unique_key"`
		Name             string                            `json:"name"`
		AuthMode         string                            `json:"auth_mode"`
		Categories       []string                          `json:"categories"`
		LogoURL          string                            `json:"logo_url,omitempty"`
		ConnectionConfig map[string]models.ConnectionField `json:"connection_config,omitempty"`
		Credentials      map[string]models.ConnectionField `json:"credentials,omitempty"`
	}
	out := make([]providerOut, 0, len(list))
	for _, p := range list {
		out = append(out, providerOut{
			UniqueKey:        p.UniqueKey,
			Name:             p.Name,
			AuthMode:         p.AuthMode,
			Categories:       p.Categories,
			LogoURL:          h.rewriteLogoURL(p.LogoURL),
			ConnectionConfig: p.ConnectionConfig,
			Credentials:      p.Credentials,
		})
	}
	writeJSON(w, http.StatusOK, out)
}

// GetProviderDocs handles GET /v1/integrations/{provider}/docs.
// Returns structured API documentation: base URL, common endpoints, code examples.
// For providers in the bundled metadata, returns rich docs. For others, falls back
// to Nango metadata with a generic usage example.
func (h *IntegrationsHandler) GetProviderDocs(w http.ResponseWriter, r *http.Request) {
	providerKey := chi.URLParam(r, "provider")

	// Tier 1: bundled metadata (rich docs for top providers).
	if doc := providers.Get(providerKey); doc != nil {
		writeJSON(w, http.StatusOK, doc)
		return
	}

	// Tier 2: check the static catalog for the display name and auth mode,
	// then try Nango for a docs URL. Never fail if Nango is unavailable.
	var (
		name     = providerKey
		authMode = "OAUTH2"
		docsURL  = ""
	)
	for _, c := range providers.Catalog {
		if c.UniqueKey == providerKey {
			name = c.Name
			authMode = c.AuthMode
			break
		}
	}
	if name == providerKey {
		// Not in catalog at all — genuinely unknown.
		writeError(w, http.StatusNotFound, "unknown provider: "+providerKey)
		return
	}

	// Best-effort Nango lookup for docs URL — ignore errors.
	if np, err := h.nango.GetProvider(r.Context(), providerKey); err == nil {
		docsURL = np.DocsURL
	} else {
		h.log.Debug("nango GetProvider unavailable (non-fatal)", zap.String("provider", providerKey), zap.Error(err))
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"provider":         providerKey,
		"name":             name,
		"base_url":         "",
		"docs_url":         docsURL,
		"auth_mode":        authMode,
		"common_endpoints": []any{},
		"note":             "Full endpoint list not bundled. Refer to the provider's official API docs.",
		"bun_example": fmt.Sprintf(
			"import { integration } from '@velane/integrations'\nconst client = integration('%s')",
			providerKey,
		),
		"python_example": fmt.Sprintf(
			"from velane.integrations import integration\nclient = integration(\"%s\")",
			providerKey,
		),
	})
}

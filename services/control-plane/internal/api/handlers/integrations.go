package handlers

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/abskrj/velane/services/control-plane/internal/models"
	"github.com/abskrj/velane/services/control-plane/internal/nango"
	"github.com/abskrj/velane/services/control-plane/internal/providers"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// IntegrationsHandler serves the provider catalog and per-provider documentation.
type IntegrationsHandler struct {
	nango            *nango.Client
	log              *zap.Logger
	nangoInternalURL string // used to proxy Nango assets through control plane
	nangoApiURL      string // browser-accessible Nango API URL, used to derive OAuth callback URL
	mcpPublicURL     string // public MCP endpoint URL shown in admin UI
}

func NewIntegrationsHandler(nangoClient *nango.Client, log *zap.Logger, nangoInternalURL, nangoApiURL, mcpPublicURL string) *IntegrationsHandler {
	return &IntegrationsHandler{
		nango:            nangoClient,
		log:              log,
		nangoInternalURL: nangoInternalURL,
		nangoApiURL:      nangoApiURL,
		mcpPublicURL:     strings.TrimSpace(mcpPublicURL),
	}
}

// ConnectInfo handles GET /v1/connect/info.
// Returns public connection metadata — currently just the OAuth callback URL that
// operators must register in their OAuth app settings (e.g. GitHub, Salesforce).
// No authentication required.
func (h *IntegrationsHandler) ConnectInfo(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"oauth_callback_url": h.nangoApiURL + "/oauth/callback",
	})
}

// MCPInfo handles GET /v1/mcp/info.
// Returns the public MCP endpoint URL for IDE configuration.
// No authentication required.
func (h *IntegrationsHandler) MCPInfo(w http.ResponseWriter, r *http.Request) {
	if h.mcpPublicURL == "" {
		writeError(w, http.StatusServiceUnavailable, "MCP URL not configured")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"mcp_url": h.mcpPublicURL,
	})
}

// rewriteLogoURL replaces Nango's self-reported logo URL with a path served
// through the control plane proxy so the browser never reaches Nango directly.
// The original URL is base64-encoded in the path so ProxyAsset knows where to
// fetch the asset from.
//
// Nango reports http:// logo URLs using its public NANGO_SERVER_URL (e.g.
// http://localhost:3003), which is unreachable inside Docker. We replace the
// host with nangoInternalURL so the proxy fetches from the correct container.
// https:// URLs (CDN assets) are kept verbatim.
func (h *IntegrationsHandler) rewriteLogoURL(nangoLogoURL string) string {
	if nangoLogoURL == "" {
		return ""
	}
	fetchURL := nangoLogoURL
	if strings.HasPrefix(nangoLogoURL, "http://") {
		rest := strings.TrimPrefix(nangoLogoURL, "http://")
		if idx := strings.Index(rest, "/"); idx >= 0 {
			fetchURL = h.nangoInternalURL + rest[idx:]
		}
	}
	encoded := base64.RawURLEncoding.EncodeToString([]byte(fetchURL))
	return "/api/v1/nango-assets/" + encoded
}

// ProxyAsset handles GET /v1/nango-assets/* and proxies static assets
// (logos, icons) through the control plane so the browser never reaches Nango.
// No auth required — these are public image files.
//
// The path segment is a base64url-encoded original URL (new format). Legacy
// path-based requests fall back to fetching from the internal Nango server.
func (h *IntegrationsHandler) ProxyAsset(w http.ResponseWriter, r *http.Request) {
	assetPath := chi.URLParam(r, "*")

	var fetchURL string
	if decoded, err := base64.RawURLEncoding.DecodeString(assetPath); err == nil &&
		(strings.HasPrefix(string(decoded), "http://") || strings.HasPrefix(string(decoded), "https://")) {
		fetchURL = string(decoded)
	} else {
		// Legacy path-based format: proxy from internal Nango server.
		fetchURL = h.nangoInternalURL + "/" + assetPath
	}

	// SSRF guard: only allow HTTPS URLs or the known internal Nango URL.
	if !strings.HasPrefix(fetchURL, "https://") && !strings.HasPrefix(fetchURL, h.nangoInternalURL) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, fetchURL, nil)
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
		DefaultScopes    []string                          `json:"default_scopes,omitempty"`
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
			DefaultScopes:    p.DefaultScopes,
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

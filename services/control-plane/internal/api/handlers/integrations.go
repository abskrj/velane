package handlers

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/abskrj/velane/services/control-plane/internal/models"
	"github.com/abskrj/velane/services/control-plane/internal/nango"
	"github.com/abskrj/velane/services/control-plane/internal/nangodocs"
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
	searchQuery := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	limit := parseProvidersLimit(r.URL.Query().Get("limit"))
	offset := parseProvidersOffset(r.URL.Query().Get("offset"))

	list, err := h.nango.ListProviders(r.Context())
	type providerOut struct {
		UniqueKey        string                            `json:"unique_key"`
		Name             string                            `json:"name"`
		AuthMode         string                            `json:"auth_mode"`
		Categories       []string                          `json:"categories"`
		DefaultScopes    []string                          `json:"default_scopes,omitempty"`
		DocsURL          string                            `json:"docs,omitempty"`
		LogoURL          string                            `json:"logo_url,omitempty"`
		ConnectionConfig map[string]models.ConnectionField `json:"connection_config,omitempty"`
		Credentials      map[string]models.ConnectionField `json:"credentials,omitempty"`
	}

	out := make([]providerOut, 0)
	if err != nil {
		h.log.Warn("nango /providers unavailable, falling back to bundled catalog", zap.Error(err))
		out = make([]providerOut, 0, len(providers.Catalog))
		for _, p := range providers.Catalog {
			out = append(out, providerOut{
				UniqueKey:  p.UniqueKey,
				Name:       p.Name,
				AuthMode:   p.AuthMode,
				Categories: p.Categories,
				DocsURL:    providers.DocsURL(p.UniqueKey),
			})
		}
	} else {
		// Normalise to a stable API shape. Include connection_config and credentials
		// so the frontend can build dynamic configuration forms.
		out = make([]providerOut, 0, len(list))
		for _, p := range list {
			docsURL := p.DocsURL
			if docsURL == "" {
				docsURL = providers.DocsURL(p.UniqueKey)
			}
			out = append(out, providerOut{
				UniqueKey:        p.UniqueKey,
				Name:             p.Name,
				AuthMode:         p.AuthMode,
				Categories:       p.Categories,
				DefaultScopes:    p.DefaultScopes,
				DocsURL:          docsURL,
				LogoURL:          h.rewriteLogoURL(p.LogoURL),
				ConnectionConfig: p.ConnectionConfig,
				Credentials:      p.Credentials,
			})
		}
	}

	filtered := make([]providerOut, 0, len(out))
	if searchQuery == "" {
		filtered = append(filtered, out...)
	} else {
		for _, p := range out {
			if strings.Contains(strings.ToLower(p.Name), searchQuery) || strings.Contains(strings.ToLower(p.UniqueKey), searchQuery) {
				filtered = append(filtered, p)
			}
		}
	}
	if offset > 0 {
		if offset >= len(filtered) {
			filtered = []providerOut{}
		} else {
			filtered = filtered[offset:]
		}
	}
	if limit > 0 && len(filtered) > limit {
		filtered = filtered[:limit]
	}
	writeJSON(w, http.StatusOK, filtered)
}

func parseProvidersLimit(raw string) int {
	if raw == "" {
		return 0
	}
	v, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || v <= 0 {
		return 0
	}
	// Keep a sane upper bound for public endpoint requests.
	if v > 100 {
		return 100
	}
	return v
}

func parseProvidersOffset(raw string) int {
	if raw == "" {
		return 0
	}
	v, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || v < 0 {
		return 0
	}
	return v
}

// GetProviderDocs handles GET /v1/integrations/{provider}/docs.
// Returns structured API documentation merged from Velane bundled metadata (when available),
// Nango provider metadata (docs URL, proxy base_url), and markdown fetched from nango.dev.
func (h *IntegrationsHandler) GetProviderDocs(w http.ResponseWriter, r *http.Request) {
	rawKey := chi.URLParam(r, "provider")
	providerKey, inCatalog := providers.ResolveKey(rawKey)

	doc := providers.Get(providerKey)
	if doc == nil {
		doc = &providers.ProviderDoc{
			Provider:        providerKey,
			CommonEndpoints: []providers.Endpoint{},
			BunExample: fmt.Sprintf(
				"import { integration } from '@velane/integrations'\nconst client = integration('%s')",
				providerKey,
			),
			PythonExample: fmt.Sprintf(
				"from velane.integrations import integration\nclient = integration(\"%s\")",
				providerKey,
			),
			Note: "Full endpoint list not bundled. See nango_docs_markdown for Nango's provider guide.",
		}
	}

	if inCatalog {
		for _, c := range providers.Catalog {
			if c.UniqueKey == providerKey {
				if doc.Name == "" {
					doc.Name = c.Name
				}
				if doc.AuthMode == "" {
					doc.AuthMode = c.AuthMode
				}
				break
			}
		}
	}

	nangoFound := h.mergeNangoProvider(r.Context(), doc)

	if doc.DocsURL == "" {
		doc.DocsURL = providers.DocsURL(providerKey)
	}

	if md, err := nangodocs.FetchMarkdown(r.Context(), doc.DocsURL, providerKey); err == nil {
		doc.NangoDocsMarkdown = md
	} else {
		h.log.Debug("nango docs markdown fetch failed (non-fatal)",
			zap.String("provider", providerKey),
			zap.Error(err),
		)
	}

	if !inCatalog && !nangoFound && doc.NangoDocsMarkdown == "" && len(doc.CommonEndpoints) == 0 {
		writeError(w, http.StatusNotFound, "unknown provider: "+rawKey)
		return
	}

	writeJSON(w, http.StatusOK, doc)
}

func (h *IntegrationsHandler) mergeNangoProvider(ctx context.Context, doc *providers.ProviderDoc) bool {
	np, err := h.nango.GetProviderDetail(ctx, doc.Provider)
	if err != nil {
		np, err = h.nango.GetProvider(ctx, doc.Provider)
		if err != nil {
			h.log.Debug("nango provider lookup unavailable (non-fatal)",
				zap.String("provider", doc.Provider),
				zap.Error(err),
			)
			return false
		}
	}

	if np.Name != "" {
		doc.Name = np.Name
	}
	if np.AuthMode != "" {
		doc.AuthMode = np.AuthMode
	}
	if np.DocsURL != "" {
		doc.DocsURL = np.DocsURL
	}
	if base := np.APIBaseURL(); base != "" {
		doc.BaseURL = base
	}
	return true
}

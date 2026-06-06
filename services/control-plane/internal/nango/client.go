package nango

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/abskrj/velane/services/control-plane/internal/models"
)

// Client is a minimal Nango REST API client.
type Client struct {
	baseURL    string
	secretKey  string
	httpClient *http.Client

	providersMu       sync.RWMutex
	providersCache    []models.NangoProvider
	providersCachedAt time.Time
}

// New returns a Nango client pointing at baseURL (e.g. "http://nango:3003").
func New(baseURL, secretKey string) *Client {
	return &Client{
		baseURL:    baseURL,
		secretKey:  secretKey,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// CreateConnectSession asks Nango for a short-lived Connect session token.
// providerConfigKey identifies a tenant-specific integration config in Nango.
func (c *Client) CreateConnectSession(ctx context.Context, tenantID, tenantName, providerConfigKey, alias string) (string, error) {
	if alias == "" {
		alias = "default"
	}
	body := map[string]any{
		// tags replaces the deprecated end_user field.
		"tags": map[string]string{
			"end_user_id":     tenantID,
			"organization_id": tenantID,
			"display_name":    tenantName,
			"velane_alias":    alias,
		},
		"allowed_integrations": []string{providerConfigKey},
	}

	b, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/connect/sessions", bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	c.setAuth(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("nango CreateConnectSession: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("nango CreateConnectSession %d: %s", resp.StatusCode, raw)
	}

	var out struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("nango CreateConnectSession decode: %w", err)
	}
	return out.Data.Token, nil
}

// PatchConnectionMetadata updates metadata on an existing Nango connection without
// overwriting fields not included in the patch (PATCH semantics).
func (c *Client) PatchConnectionMetadata(ctx context.Context, connectionID, providerConfigKey string, metadata map[string]any) error {
	body := map[string]any{
		"connection_id":       connectionID,
		"provider_config_key": providerConfigKey,
		"metadata":            metadata,
	}
	b, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, c.baseURL+"/connections/metadata", bytes.NewReader(b))
	if err != nil {
		return err
	}
	c.setAuth(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("nango PatchConnectionMetadata: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("nango PatchConnectionMetadata %d: %s", resp.StatusCode, raw)
	}
	return nil
}

// DeleteConnection removes a connection from Nango.
func (c *Client) DeleteConnection(ctx context.Context, connectionID, provider string) error {
	url := fmt.Sprintf("%s/connection/%s?provider_config_key=%s", c.baseURL, connectionID, provider)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	c.setAuth(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("nango DeleteConnection: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 && resp.StatusCode != 404 {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("nango DeleteConnection %d: %s", resp.StatusCode, raw)
	}
	return nil
}

// ListProviders returns all provider templates from Nango, cached for 1 hour.
func (c *Client) ListProviders(ctx context.Context) ([]models.NangoProvider, error) {
	c.providersMu.RLock()
	if time.Since(c.providersCachedAt) < time.Hour && c.providersCache != nil {
		cached := c.providersCache
		c.providersMu.RUnlock()
		return cached, nil
	}
	c.providersMu.RUnlock()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/providers", nil)
	if err != nil {
		return nil, err
	}
	c.setAuth(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("nango ListProviders: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("nango ListProviders %d: %s", resp.StatusCode, raw)
	}

	// Nango wraps the list in {"data": [...]}
	var envelope struct {
		Data []models.NangoProvider `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, fmt.Errorf("nango ListProviders decode: %w", err)
	}
	providers := envelope.Data

	c.providersMu.Lock()
	c.providersCache = providers
	c.providersCachedAt = time.Now()
	c.providersMu.Unlock()

	return providers, nil
}

// GetProvider returns metadata for a single provider from the cached list.
func (c *Client) GetProvider(ctx context.Context, providerKey string) (*models.NangoProvider, error) {
	providers, err := c.ListProviders(ctx)
	if err != nil {
		return nil, err
	}
	for i := range providers {
		if providers[i].UniqueKey == providerKey { // UniqueKey maps to Nango's "name" field (the slug)
			return &providers[i], nil
		}
	}
	return nil, fmt.Errorf("provider %q not found", providerKey)
}

// Proxy forwards an HTTP request to Nango's proxy endpoint, which injects
// the OAuth token and proxies to the external provider API.
// path is the provider-relative path, e.g. "/user/repos".
func (c *Client) Proxy(w http.ResponseWriter, r *http.Request, connectionID, providerConfigKey, path string) {
	// Build the Nango proxy URL: strip leading slash from path then concatenate.
	nangoURL := c.baseURL + "/proxy" + path
	if r.URL.RawQuery != "" {
		nangoURL += "?" + r.URL.RawQuery
	}

	proxyReq, err := http.NewRequestWithContext(r.Context(), r.Method, nangoURL, r.Body)
	if err != nil {
		http.Error(w, "proxy request build failed", http.StatusInternalServerError)
		return
	}

	// Forward content-type if set.
	if ct := r.Header.Get("Content-Type"); ct != "" {
		proxyReq.Header.Set("Content-Type", ct)
	}

	c.setAuth(proxyReq)
	proxyReq.Header.Set("Connection-Id", connectionID)
	proxyReq.Header.Set("Provider-Config-Key", providerConfigKey)

	resp, err := c.httpClient.Do(proxyReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("proxy failed: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Forward response headers and status.
	for k, vs := range resp.Header {
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body) //nolint:errcheck
}

// CreateIntegrationConfig creates a provider config in Nango.
// Nango API: POST /integrations with unique_key, provider, credentials.
func (c *Client) CreateIntegrationConfig(ctx context.Context, providerConfigKey, provider string, credentials map[string]any) error {
	body := map[string]any{
		"unique_key": providerConfigKey,
		"provider":   provider,
	}
	if len(credentials) > 0 {
		body["credentials"] = credentials
	}

	b, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/integrations", bytes.NewReader(b))
	if err != nil {
		return err
	}
	c.setAuth(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("nango CreateIntegrationConfig: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("nango CreateIntegrationConfig %d: %s", resp.StatusCode, raw)
	}
	return nil
}

// UpdateIntegrationConfig updates an existing provider config in Nango.
// Nango API: PATCH /integrations/{uniqueKey}
func (c *Client) UpdateIntegrationConfig(ctx context.Context, providerConfigKey string, credentials map[string]any) error {
	body := map[string]any{}
	if len(credentials) > 0 {
		body["credentials"] = credentials
	}

	b, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, c.baseURL+"/integrations/"+providerConfigKey, bytes.NewReader(b))
	if err != nil {
		return err
	}
	c.setAuth(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("nango UpdateIntegrationConfig: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("nango UpdateIntegrationConfig %d: %s", resp.StatusCode, raw)
	}
	return nil
}

// ListIntegrationConfigs returns all configured integrations from Nango.
// Nango API: GET /integrations → {"data":[...]}
func (c *Client) ListIntegrationConfigs(ctx context.Context) ([]models.NangoIntegrationConfig, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/integrations", nil)
	if err != nil {
		return nil, err
	}
	c.setAuth(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("nango ListIntegrationConfigs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("nango ListIntegrationConfigs %d: %s", resp.StatusCode, raw)
	}

	var envelope struct {
		Data []models.NangoIntegrationConfig `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, fmt.Errorf("nango ListIntegrationConfigs decode: %w", err)
	}
	return envelope.Data, nil
}

// DeleteIntegrationConfig removes a provider config from Nango.
// Nango API: DELETE /integrations/{uniqueKey}
func (c *Client) DeleteIntegrationConfig(ctx context.Context, providerConfigKey string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.baseURL+"/integrations/"+providerConfigKey, nil)
	if err != nil {
		return err
	}
	c.setAuth(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("nango DeleteIntegrationConfig: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 && resp.StatusCode != 404 {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("nango DeleteIntegrationConfig %d: %s", resp.StatusCode, raw)
	}
	return nil
}

func (c *Client) setAuth(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.secretKey)
}

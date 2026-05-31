package providers

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed providers_meta.json
var rawMeta []byte

// Endpoint describes one API endpoint for a provider.
type Endpoint struct {
	Method      string         `json:"method"`
	Path        string         `json:"path"`
	Description string         `json:"description"`
	Body        map[string]any `json:"body,omitempty"`
	Query       map[string]any `json:"query,omitempty"`
}

// ProviderDoc is the full documentation entry for one provider.
type ProviderDoc struct {
	Provider        string     `json:"provider"`
	Name            string     `json:"name"`
	BaseURL         string     `json:"base_url"`
	DocsURL         string     `json:"docs_url"`
	AuthMode        string     `json:"auth_mode"`
	CommonEndpoints []Endpoint `json:"common_endpoints"`
	BunExample      string     `json:"bun_example"`
	PythonExample   string     `json:"python_example"`
	Note            string     `json:"note,omitempty"`
}

type meta struct {
	Name            string     `json:"name"`
	BaseURL         string     `json:"base_url"`
	DocsURL         string     `json:"docs_url"`
	AuthMode        string     `json:"auth_mode"`
	CommonEndpoints []Endpoint `json:"common_endpoints"`
	BunExample      string     `json:"bun_example"`
	PythonExample   string     `json:"python_example"`
}

var bundled map[string]meta

func init() {
	if err := json.Unmarshal(rawMeta, &bundled); err != nil {
		panic(fmt.Sprintf("providers: failed to parse providers_meta.json: %v", err))
	}
}

// Get returns the bundled ProviderDoc for the given provider slug, or nil if not bundled.
func Get(providerKey string) *ProviderDoc {
	m, ok := bundled[providerKey]
	if !ok {
		return nil
	}
	return &ProviderDoc{
		Provider:        providerKey,
		Name:            m.Name,
		BaseURL:         m.BaseURL,
		DocsURL:         m.DocsURL,
		AuthMode:        m.AuthMode,
		CommonEndpoints: m.CommonEndpoints,
		BunExample:      m.BunExample,
		PythonExample:   m.PythonExample,
	}
}

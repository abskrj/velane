package models

import "time"

// Connection represents an OAuth integration a tenant has connected via Nango.
type Connection struct {
	ID                  string    `json:"id"`
	TenantID            string    `json:"tenant_id"`
	Provider            string    `json:"provider"`
	Alias               string    `json:"alias"`
	ProviderConfigKey   string    `json:"provider_config_key"`
	CredentialProfileID *string   `json:"credential_profile_id,omitempty"`
	NangoConnectionID   string    `json:"nango_connection_id"`
	DisplayName         string    `json:"display_name,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// NangoProvider is metadata about a single provider returned by Nango's /providers endpoint.
// Nango's field naming: `name` is the slug/unique key, `display_name` is human-readable.
type NangoProvider struct {
	UniqueKey        string                     `json:"name"`         // slug e.g. "github"
	Name             string                     `json:"display_name"` // e.g. "GitHub"
	AuthMode         string                     `json:"auth_mode"`
	Categories       []string                   `json:"categories"`
	DefaultScopes    []string                   `json:"default_scopes,omitempty"`
	DocsURL          string                     `json:"docs,omitempty"`
	LogoURL          string                     `json:"logo_url,omitempty"`
	ConnectionConfig map[string]ConnectionField `json:"connection_config,omitempty"`
	Credentials      map[string]ConnectionField `json:"credentials,omitempty"`
}

// ConnectionField describes one field in a provider's connection_config or credentials schema.
type ConnectionField struct {
	Type        string `json:"type"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Example     string `json:"example,omitempty"`
	Optional    bool   `json:"optional,omitempty"`
	Automated   bool   `json:"automated,omitempty"`
	Prefix      string `json:"prefix,omitempty"`
}

// NangoIntegrationConfig is a configured integration stored in Nango.
// Matches Nango's GET /integrations response shape.
type NangoIntegrationConfig struct {
	UniqueKey   string `json:"unique_key"`
	Provider    string `json:"provider"`
	DisplayName string `json:"display_name,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
}

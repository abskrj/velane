package models

import "time"

// IntegrationCredentialProfile stores tenant-scoped provider credentials and metadata.
// Secret fields are encrypted inside config_json and never returned in plaintext.
type IntegrationCredentialProfile struct {
	ID                     string    `json:"id"`
	TenantID               string    `json:"tenant_id"`
	Provider               string    `json:"provider"`
	Alias                  string    `json:"alias"`
	Name                   string    `json:"name"`
	NangoProviderConfigKey string    `json:"nango_provider_config_key"`
	ConfigJSON             []byte    `json:"-"`
	IsDefault              bool      `json:"is_default"`
	CreatedBy              string    `json:"created_by"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

// IntegrationCredentialProfileView is the API-safe representation returned to clients.
type IntegrationCredentialProfileView struct {
	ID                     string    `json:"id"`
	TenantID               string    `json:"tenant_id"`
	Provider               string    `json:"provider"`
	Alias                  string    `json:"alias"`
	Name                   string    `json:"name"`
	NangoProviderConfigKey string    `json:"nango_provider_config_key"`
	CredentialsType        string    `json:"credentials_type"`
	OAuthScopes            string    `json:"oauth_scopes,omitempty"`
	IsDefault              bool      `json:"is_default"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

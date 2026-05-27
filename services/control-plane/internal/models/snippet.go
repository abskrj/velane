package models

import "time"

// Language identifies which executor runtime should be used.
type Language string

const (
	LanguageBun    Language = "bun"
	LanguagePython Language = "python"
)

// Snippet is a named, versioned unit of user code belonging to a tenant.
type Snippet struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Language  Language  `json:"language"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy string    `json:"created_by"`
}

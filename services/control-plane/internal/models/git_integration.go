package models

import "time"

// GitIntegration links a snippet to a git provider for push-to-deploy webhooks.
type GitIntegration struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	SnippetID string    `json:"snippet_id"`
	Provider  string    `json:"provider"`
	RepoURL   string    `json:"repo_url"`
	Secret    string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

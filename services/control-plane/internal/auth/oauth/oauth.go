// Package oauth implements social login (Google, GitHub) for the admin portal.
// It is entirely separate from Nango, which handles tenant integration OAuth.
package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"golang.org/x/oauth2"
)

// UserInfo is the normalized identity returned by a provider after token exchange.
type UserInfo struct {
	Subject       string // stable provider-issued user id
	Email         string
	EmailVerified bool
}

// Provider exchanges OAuth codes and fetches the authenticated user's profile.
type Provider struct {
	name     string
	config   *oauth2.Config
	fetch    func(ctx context.Context, token *oauth2.Token) (*UserInfo, error)
	authOpts []oauth2.AuthCodeOption
}

// Name returns the provider slug (e.g. "google").
func (p *Provider) Name() string { return p.name }

// AuthCodeURL builds the provider authorization URL for the given state.
func (p *Provider) AuthCodeURL(state string) string {
	return p.config.AuthCodeURL(state, p.authOpts...)
}

// Exchange swaps an authorization code for a token and fetches the user profile.
func (p *Provider) Exchange(ctx context.Context, code string) (*UserInfo, error) {
	token, err := p.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("token exchange: %w", err)
	}
	return p.fetch(ctx, token)
}

// Manager holds the configured providers keyed by slug.
type Manager struct {
	publicBaseURL string
	providers     map[string]*Provider
}

// NewManager creates an empty provider manager. publicBaseURL is the browser-facing
// admin origin; OAuth callbacks are served under publicBaseURL + "/api".
func NewManager(publicBaseURL string) *Manager {
	return &Manager{
		publicBaseURL: strings.TrimRight(publicBaseURL, "/"),
		providers:     make(map[string]*Provider),
	}
}

// Get returns the provider for a slug, or nil if not configured.
func (m *Manager) Get(name string) *Provider { return m.providers[name] }

// Names returns the slugs of all configured providers.
func (m *Manager) Names() []string {
	names := make([]string, 0, len(m.providers))
	for name := range m.providers {
		names = append(names, name)
	}
	return names
}

// Len returns the number of configured providers.
func (m *Manager) Len() int { return len(m.providers) }

func (m *Manager) redirectURL(provider string) string {
	return fmt.Sprintf("%s/api/v1/admin/auth/oauth/%s/callback", m.publicBaseURL, provider)
}

// AddGoogle registers the Google provider.
func (m *Manager) AddGoogle(clientID, clientSecret string) {
	m.providers["google"] = &Provider{
		name: "google",
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  m.redirectURL("google"),
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://accounts.google.com/o/oauth2/v2/auth",
				TokenURL: "https://oauth2.googleapis.com/token",
			},
		},
		fetch: fetchGoogleUser,
	}
}

// AddGitHub registers the GitHub provider.
func (m *Manager) AddGitHub(clientID, clientSecret string) {
	m.providers["github"] = &Provider{
		name: "github",
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  m.redirectURL("github"),
			Scopes:       []string{"read:user", "user:email"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://github.com/login/oauth/authorize",
				TokenURL: "https://github.com/login/oauth/access_token",
			},
		},
		fetch: fetchGitHubUser,
	}
}

func fetchGoogleUser(ctx context.Context, token *oauth2.Token) (*UserInfo, error) {
	body, err := getJSON(ctx, token, "https://openidconnect.googleapis.com/v1/userinfo", nil)
	if err != nil {
		return nil, err
	}
	var u struct {
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
	}
	if err := json.Unmarshal(body, &u); err != nil {
		return nil, fmt.Errorf("decode google userinfo: %w", err)
	}
	if u.Sub == "" {
		return nil, fmt.Errorf("google userinfo missing subject")
	}
	return &UserInfo{Subject: u.Sub, Email: strings.ToLower(u.Email), EmailVerified: u.EmailVerified}, nil
}

func fetchGitHubUser(ctx context.Context, token *oauth2.Token) (*UserInfo, error) {
	headers := map[string]string{"Accept": "application/vnd.github+json"}

	body, err := getJSON(ctx, token, "https://api.github.com/user", headers)
	if err != nil {
		return nil, err
	}
	var profile struct {
		ID    int64  `json:"id"`
		Email string `json:"email"`
	}
	if err := json.Unmarshal(body, &profile); err != nil {
		return nil, fmt.Errorf("decode github user: %w", err)
	}
	if profile.ID == 0 {
		return nil, fmt.Errorf("github user missing id")
	}

	info := &UserInfo{Subject: fmt.Sprintf("%d", profile.ID)}

	// GitHub's /user endpoint often omits the email (or returns an unverified one);
	// the authoritative source is /user/emails.
	emailsBody, err := getJSON(ctx, token, "https://api.github.com/user/emails", headers)
	if err == nil {
		var emails []struct {
			Email    string `json:"email"`
			Primary  bool   `json:"primary"`
			Verified bool   `json:"verified"`
		}
		if json.Unmarshal(emailsBody, &emails) == nil {
			for _, e := range emails {
				if e.Primary && e.Verified {
					info.Email = strings.ToLower(e.Email)
					info.EmailVerified = true
					break
				}
			}
			if info.Email == "" {
				for _, e := range emails {
					if e.Verified {
						info.Email = strings.ToLower(e.Email)
						info.EmailVerified = true
						break
					}
				}
			}
		}
	}

	if info.Email == "" && profile.Email != "" {
		info.Email = strings.ToLower(profile.Email)
	}
	return info, nil
}

func getJSON(ctx context.Context, token *oauth2.Token, url string, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	token.SetAuthHeader(req)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", url, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: status %d", url, resp.StatusCode)
	}
	return body, nil
}

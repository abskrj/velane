package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/abskrj/velane/services/control-plane/internal/auth"
	"github.com/abskrj/velane/services/control-plane/internal/models"
	"go.uber.org/zap"
)

type sessionContextKey string

const (
	sessionUserKey    sessionContextKey = "session_user"
	sessionRoleKey    sessionContextKey = "session_role"
	SessionCookieName                   = "velane_session"
	ActiveOrgCookieName                 = "velane_active_org"
)

// SessionStore is the subset of the store needed by SessionAuth to resolve tenant membership.
type SessionStore interface {
	ListUserTenantMemberships(ctx context.Context, userID string) ([]*models.UserTenantMembership, error)
}

// SessionAuth validates a Bearer JWT token (from admin portal login) and sets the user in
// context. It resolves the user's active org membership (from cookie or first membership)
// and stores tenant + role in context so RequireScope can grant access without an API key.
// Falls through to next if the token is absent or invalid, allowing API key auth to handle it.
func SessionAuth(provider auth.Provider, store SessionStore, log *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw, ok := sessionTokenFromRequest(r)
			if !ok {
				next.ServeHTTP(w, r)
				return
			}
			user, err := provider.ValidateSession(r.Context(), raw)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), sessionUserKey, user)

			if store != nil {
				if memberships, err := store.ListUserTenantMemberships(ctx, user.ID); err == nil && len(memberships) > 0 {
					selected := memberships[0]
					if cookie, err := r.Cookie(ActiveOrgCookieName); err == nil {
						cookieSlug := strings.TrimSpace(cookie.Value)
						if cookieSlug != "" {
							for _, membership := range memberships {
								if membership.Slug == cookieSlug {
									selected = membership
									break
								}
							}
						}
					}
					ctx = context.WithValue(ctx, sessionRoleKey, selected.Role)
					ctx = context.WithValue(ctx, tenantKey, &models.Tenant{
						ID:   selected.TenantID,
						Slug: selected.Slug,
						Name: selected.Name,
					})
				}
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func sessionTokenFromRequest(r *http.Request) (string, bool) {
	header := r.Header.Get("Authorization")
	if strings.HasPrefix(header, "Bearer ") {
		raw := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
		if raw != "" {
			return raw, true
		}
	}

	cookie, err := r.Cookie(SessionCookieName)
	if err != nil || strings.TrimSpace(cookie.Value) == "" {
		return "", false
	}
	return strings.TrimSpace(cookie.Value), true
}

// SessionUserFromContext retrieves the authenticated session user from the request context.
func SessionUserFromContext(ctx context.Context) *models.User {
	u, _ := ctx.Value(sessionUserKey).(*models.User)
	return u
}

// SessionRoleFromContext retrieves the tenant-scoped role for the session user.
func SessionRoleFromContext(ctx context.Context) string {
	r, _ := ctx.Value(sessionRoleKey).(string)
	return r
}

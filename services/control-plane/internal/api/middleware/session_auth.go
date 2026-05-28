package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/runeforge/control-plane/internal/auth"
	"github.com/runeforge/control-plane/internal/models"
	"go.uber.org/zap"
)

type sessionContextKey string

const sessionUserKey sessionContextKey = "session_user"

// SessionAuth validates a Bearer session token (from web login) and sets the user in context.
// Falls through to next if no Authorization header — can be chained after API key auth.
func SessionAuth(provider auth.Provider, log *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				next.ServeHTTP(w, r)
				return
			}
			raw := strings.TrimPrefix(header, "Bearer ")
			user, err := provider.ValidateSession(r.Context(), raw)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			ctx := context.WithValue(r.Context(), sessionUserKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// SessionUserFromContext retrieves the authenticated session user from the request context.
func SessionUserFromContext(ctx context.Context) *models.User {
	u, _ := ctx.Value(sessionUserKey).(*models.User)
	return u
}

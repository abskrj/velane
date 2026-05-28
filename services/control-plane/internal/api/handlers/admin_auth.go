package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/runeforge/control-plane/internal/api/middleware"
	"github.com/runeforge/control-plane/internal/auth"
	"github.com/runeforge/control-plane/internal/models"
	"go.uber.org/zap"
)

// AdminAuthStore is the subset of *postgres.Store that admin auth handlers need.
type AdminAuthStore interface {
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetInviteByTokenHash(ctx context.Context, hash string) (*models.InviteToken, error)
	AcceptInvite(ctx context.Context, id string) error
	AddMember(ctx context.Context, tenantID, userID, role string) (*models.TenantMember, error)
}

// AdminAuthHandler handles email/password auth for the admin portal.
type AdminAuthHandler struct {
	provider auth.Provider
	store    AdminAuthStore
	log      *zap.Logger
}

// NewAdminAuthHandler constructs an AdminAuthHandler.
func NewAdminAuthHandler(provider auth.Provider, store AdminAuthStore, log *zap.Logger) *AdminAuthHandler {
	return &AdminAuthHandler{provider: provider, store: store, log: log}
}

type registerRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	InviteToken string `json:"invite_token,omitempty"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Register handles POST /v1/admin/auth/register.
func (h *AdminAuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	user, err := h.provider.CreateUser(r.Context(), req.Email, req.Password)
	if err != nil {
		h.log.Error("create user failed", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	// If an invite token was supplied, validate it and accept.
	if req.InviteToken != "" {
		inviteHash := hashInviteToken(req.InviteToken)
		invite, err := h.store.GetInviteByTokenHash(r.Context(), inviteHash)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid or expired invite token")
			return
		}
		if invite.AcceptedAt != nil {
			writeError(w, http.StatusBadRequest, "invite token already used")
			return
		}
		if err := h.store.AcceptInvite(r.Context(), invite.ID); err != nil {
			h.log.Error("accept invite failed", zap.Error(err))
			writeError(w, http.StatusInternalServerError, "failed to accept invite")
			return
		}
		if _, err := h.store.AddMember(r.Context(), invite.TenantID, user.ID, invite.Role); err != nil {
			h.log.Error("add member failed", zap.Error(err))
			writeError(w, http.StatusInternalServerError, "failed to add member")
			return
		}
	}

	// Authenticate immediately to return a session.
	sess, err := h.provider.Authenticate(r.Context(), req.Email, req.Password)
	if err != nil {
		h.log.Error("post-register authenticate failed", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to create session")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"user":          user,
		"session_token": sess.Token,
		"expires_at":    sess.ExpiresAt,
	})
}

// Login handles POST /v1/admin/auth/login.
func (h *AdminAuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	sess, err := h.provider.Authenticate(r.Context(), req.Email, req.Password)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"session_token": sess.Token,
		"expires_at":    sess.ExpiresAt,
	})
}

// Logout handles POST /v1/admin/auth/logout.
func (h *AdminAuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	header := r.Header.Get("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		writeError(w, http.StatusBadRequest, "missing Authorization header")
		return
	}
	raw := strings.TrimPrefix(header, "Bearer ")

	if err := h.provider.InvalidateSession(r.Context(), raw); err != nil {
		h.log.Debug("invalidate session failed", zap.Error(err))
	}

	w.WriteHeader(http.StatusNoContent)
}

// Me handles GET /v1/admin/auth/me.
func (h *AdminAuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	user := middleware.SessionUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	writeJSON(w, http.StatusOK, user)
}

// hashInviteToken hashes a raw invite token using SHA-256 (same algorithm as auth package).
func hashInviteToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

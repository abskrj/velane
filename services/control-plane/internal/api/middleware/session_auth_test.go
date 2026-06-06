package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/abskrj/velane/services/control-plane/internal/api/middleware"
	"github.com/abskrj/velane/services/control-plane/internal/models"
)

type mockSessionProvider struct {
	validateFn func(ctx context.Context, rawToken string) (*models.User, error)
}

func (m *mockSessionProvider) CreateUser(ctx context.Context, email, password string) (*models.User, error) {
	panic("unexpected CreateUser call")
}

func (m *mockSessionProvider) Authenticate(ctx context.Context, email, password string) (*models.Session, error) {
	panic("unexpected Authenticate call")
}

func (m *mockSessionProvider) ValidateSession(ctx context.Context, rawToken string) (*models.User, error) {
	return m.validateFn(ctx, rawToken)
}

func (m *mockSessionProvider) InvalidateSession(ctx context.Context, rawToken string) error {
	panic("unexpected InvalidateSession call")
}

type mockSessionStore struct {
	listUserTenantMembershipsFn func(ctx context.Context, userID string) ([]*models.UserTenantMembership, error)
}

func (m *mockSessionStore) ListUserTenantMemberships(ctx context.Context, userID string) ([]*models.UserTenantMembership, error) {
	return m.listUserTenantMembershipsFn(ctx, userID)
}

func TestSessionAuth_UsesCookieAndAttachesTenantRole(t *testing.T) {
	user := &models.User{ID: "u1", Email: "alice@example.com"}
	mw := middleware.SessionAuth(&mockSessionProvider{
		validateFn: func(_ context.Context, rawToken string) (*models.User, error) {
			if rawToken != "cookie-token" {
				t.Fatalf("rawToken = %q; want cookie-token", rawToken)
			}
			return user, nil
		},
	}, &mockSessionStore{
		listUserTenantMembershipsFn: func(_ context.Context, userID string) ([]*models.UserTenantMembership, error) {
			if userID != "u1" {
				t.Fatalf("userID = %q; want u1", userID)
			}
			return []*models.UserTenantMembership{{
				TenantID: "t1",
				Slug:     "acme",
				Name:     "Acme",
				Role:     "admin",
			}}, nil
		},
	}, nopLog)

	var gotUser *models.User
	var gotTenant *models.Tenant
	var gotRole string
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUser = middleware.SessionUserFromContext(r.Context())
		gotTenant = middleware.TenantFromContext(r.Context())
		gotRole = middleware.SessionRoleFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: middleware.SessionCookieName, Value: "cookie-token"})
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200", rec.Code)
	}
	if gotUser == nil || gotUser.ID != "u1" {
		t.Fatalf("user not attached correctly: %+v", gotUser)
	}
	if gotTenant == nil || gotTenant.ID != "t1" {
		t.Fatalf("tenant not attached correctly: %+v", gotTenant)
	}
	if gotRole != "admin" {
		t.Fatalf("role = %q; want admin", gotRole)
	}
}

func TestSessionAuth_BearerStillWorksDuringMigration(t *testing.T) {
	mw := middleware.SessionAuth(&mockSessionProvider{
		validateFn: func(_ context.Context, rawToken string) (*models.User, error) {
			if rawToken != "bearer-token" {
				t.Fatalf("rawToken = %q; want bearer-token", rawToken)
			}
			return &models.User{ID: "u1"}, nil
		},
	}, nil, nopLog)

	var gotUser *models.User
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUser = middleware.SessionUserFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer bearer-token")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200", rec.Code)
	}
	if gotUser == nil || gotUser.ID != "u1" {
		t.Fatalf("user not attached correctly: %+v", gotUser)
	}
}

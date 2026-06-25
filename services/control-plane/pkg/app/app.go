// Package app is the public extension surface for velane-cloud and other
// private overlays that build on top of the open-source velane control plane.
// It wraps internal initialisation and exposes only what cloud extensions need.
package app

import (
	"context"

	internalapp "github.com/abskrj/velane/services/control-plane/internal/app"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// App holds the initialised router and a minimal set of capabilities that
// cloud extensions can use to mount additional routes.
type App struct {
	// Router is the fully configured chi mux. Mount cloud-only routes on it
	// before passing it to an http.Server.
	Router *chi.Mux

	// Log is the shared production logger.
	Log *zap.Logger

	// Port is the value of the PORT env var (default "8080").
	Port string

	inner *internalapp.App
}

// Bootstrap initialises every dependency and returns a ready-to-extend App.
// Background goroutines run in ctx; cancel it before calling Close.
func Bootstrap(ctx context.Context, log *zap.Logger) (*App, error) {
	inner, err := internalapp.Bootstrap(ctx, log)
	if err != nil {
		return nil, err
	}
	return &App{
		Router: inner.Router,
		Log:    inner.Log,
		Port:   inner.Port,
		inner:  inner,
	}, nil
}

// SetTenantLicenseKey sets or clears the license key for a cloud tenant.
// Pass nil to clear the key (revoke cloud license).
func (a *App) SetTenantLicenseKey(ctx context.Context, tenantID string, licenseKey *string) error {
	return a.inner.Store.SetTenantLicenseKey(ctx, tenantID, licenseKey)
}

// Close shuts down long-lived resources. Call after the HTTP server has stopped.
func (a *App) Close() {
	a.inner.Close()
}

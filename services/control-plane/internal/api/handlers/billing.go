package handlers

import (
	"context"
	"net/http"

	"github.com/abskrj/velane/services/control-plane/internal/api/middleware"
	"github.com/abskrj/velane/services/control-plane/internal/license"
	"github.com/abskrj/velane/services/control-plane/internal/models"
	"go.uber.org/zap"
)

type billingTenantStore interface {
	GetTenantByID(ctx context.Context, id string) (*models.Tenant, error)
}

type BillingHandler struct {
	store  billingTenantStore
	licMgr *license.Manager
	log    *zap.Logger
}

func NewBillingHandler(store billingTenantStore, licMgr *license.Manager, log *zap.Logger) *BillingHandler {
	return &BillingHandler{store: store, licMgr: licMgr, log: log}
}

type tenantPlanResponse struct {
	Plan     string   `json:"plan"`
	Valid    bool     `json:"valid"`
	Features []string `json:"features"`
}

// GetPlan handles GET /v1/tenant/plan.
// Returns the plan, validity, and active features for the authenticated tenant.
func (h *BillingHandler) GetPlan(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.TenantFromContext(r.Context())
	if tenant == nil {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	// Session auth sets a minimal tenant without license_key — do a fresh DB lookup.
	full, err := h.store.GetTenantByID(r.Context(), tenant.ID)
	if err != nil {
		h.log.Error("billing: get tenant", zap.String("tenant_id", tenant.ID), zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to load plan")
		return
	}

	key := ""
	if full.LicenseKey != nil {
		key = *full.LicenseKey
	}

	plan, features, valid := h.licMgr.TenantStatus(r.Context(), key)
	if features == nil {
		features = []string{}
	}

	writeJSON(w, http.StatusOK, tenantPlanResponse{
		Plan:     plan,
		Valid:    valid,
		Features: features,
	})
}

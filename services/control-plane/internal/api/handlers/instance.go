package handlers

import (
	"net/http"
	"os"

	"github.com/abskrj/velane/services/control-plane/internal/license"
	"go.uber.org/zap"
)

type InstanceHandler struct {
	licMgr    *license.Manager
	cloudMode bool
	log       *zap.Logger
}

func NewInstanceHandler(licMgr *license.Manager, log *zap.Logger) *InstanceHandler {
	return &InstanceHandler{
		licMgr:    licMgr,
		cloudMode: os.Getenv("VELANE_CLOUD") == "true",
		log:       log,
	}
}

type instanceInfoResponse struct {
	Cloud        bool     `json:"cloud"`
	Plan         string   `json:"plan"`
	LicenseValid bool     `json:"license_valid"`
	Features     []string `json:"features"`
}

func (h *InstanceHandler) GetInfo(w http.ResponseWriter, r *http.Request) {
	plan, features, valid := h.licMgr.InstanceStatus(r.Context())
	if features == nil {
		features = []string{}
	}
	writeJSON(w, http.StatusOK, instanceInfoResponse{
		Cloud:        h.cloudMode,
		Plan:         plan,
		LicenseValid: valid,
		Features:     features,
	})
}

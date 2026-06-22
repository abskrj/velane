package handlers

import (
	"net/http"

	"github.com/abskrj/velane/services/control-plane/internal/license"
	"go.uber.org/zap"
)

type InstanceHandler struct {
	licMgr *license.Manager
	log    *zap.Logger
}

func NewInstanceHandler(licMgr *license.Manager, log *zap.Logger) *InstanceHandler {
	return &InstanceHandler{licMgr: licMgr, log: log}
}

type instanceInfoResponse struct {
	Features []string `json:"features"`
}

func (h *InstanceHandler) GetInfo(w http.ResponseWriter, r *http.Request) {
	features := h.licMgr.InstanceFeatures(r.Context())
	if features == nil {
		features = []string{}
	}
	writeJSON(w, http.StatusOK, instanceInfoResponse{Features: features})
}

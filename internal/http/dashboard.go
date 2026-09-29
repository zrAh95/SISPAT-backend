package httpapi

import (
	"net/http"
	"strconv"

	"sispat/backend/internal/auth"
	"sispat/backend/internal/dashboard"
)

type DashboardHandler struct{ service *dashboard.Service }

func NewDashboardHandler(service *dashboard.Service) *DashboardHandler {
	return &DashboardHandler{service: service}
}

func (h *DashboardHandler) Summary(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserFromContext(r.Context())
	data, err := h.service.Summary(user.PenggunaID, user.Roles)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "gagal mengambil dashboard")
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *DashboardHandler) Kehadiran(w http.ResponseWriter, r *http.Request) {
	data, err := h.service.KehadiranPerRegu()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "gagal mengambil monitoring regu")
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *DashboardHandler) Rekap(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserFromContext(r.Context())
	reguID, _ := strconv.ParseUint(r.URL.Query().Get("regu_id"), 10, 64)
	data, err := h.service.Rekap(user.PenggunaID, user.Roles, r.URL.Query().Get("tanggal_awal"), r.URL.Query().Get("tanggal_akhir"), uint(reguID))
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "gagal mengambil rekap")
		return
	}
	writeJSON(w, http.StatusOK, data)
}

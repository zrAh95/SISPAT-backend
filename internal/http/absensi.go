package httpapi

import (
	"net/http"

	"sispat/backend/internal/absensi"
	"sispat/backend/internal/auth"
	"sispat/backend/internal/config"
)

type AbsensiHandler struct {
	service *absensi.Service
	cfg     config.Config
}

func NewAbsensiHandler(service *absensi.Service, cfg config.Config) *AbsensiHandler {
	return &AbsensiHandler{service: service, cfg: cfg}
}

func (h *AbsensiHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "silakan masuk terlebih dahulu")
		return
	}
	data, err := h.service.List(user.PenggunaID, user.Roles)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "gagal mengambil sesi ronda")
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *AbsensiHandler) Start(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "silakan masuk terlebih dahulu")
		return
	}
	id, ok := parsePathID(w, r, "id")
	if !ok {
		return
	}
	if err := h.service.Start(id, user.PenggunaID, user.Roles); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"pesan": "sesi ronda dimulai"})
}

func (h *AbsensiHandler) Create(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "silakan masuk terlebih dahulu")
		return
	}
	id, ok := parsePathID(w, r, "id")
	if !ok {
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, h.cfg.MaxUploadBytes+(1024*1024))
	if err := r.ParseMultipartForm(h.cfg.MaxUploadBytes); err != nil {
		writeJSONError(w, http.StatusBadRequest, "file foto tidak valid")
		return
	}
	file, header, err := r.FormFile("selfie")
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "selfie wajib diunggah")
		return
	}
	defer file.Close()

	result, err := h.service.Create(id, user.PenggunaID, user.Roles, file, header)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

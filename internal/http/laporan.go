package httpapi

import (
	"encoding/json"
	"net/http"

	"sispat/backend/internal/auth"
	"sispat/backend/internal/config"
	"sispat/backend/internal/laporan"
)

type LaporanHandler struct {
	service *laporan.Service
	cfg     config.Config
}

func NewLaporanHandler(service *laporan.Service, cfg config.Config) *LaporanHandler {
	return &LaporanHandler{service: service, cfg: cfg}
}

func (h *LaporanHandler) Create(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserFromContext(r.Context())
	sessionID, ok := parsePathID(w, r, "id")
	if !ok {
		return
	}
	var input laporan.CreateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSONError(w, http.StatusBadRequest, "format permintaan tidak valid")
		return
	}
	data, err := h.service.Create(sessionID, user.PenggunaID, user.Roles, input)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, data)
}

func (h *LaporanHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePathID(w, r, "id")
	if !ok {
		return
	}
	data, err := h.service.Get(id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "laporan tidak ditemukan")
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *LaporanHandler) Evidence(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserFromContext(r.Context())
	id, ok := parsePathID(w, r, "id")
	if !ok {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, h.cfg.MaxUploadBytes+(1024*1024))
	if err := r.ParseMultipartForm(h.cfg.MaxUploadBytes); err != nil {
		writeJSONError(w, http.StatusBadRequest, "file evidence tidak valid")
		return
	}
	file, header, err := r.FormFile("bukti")
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "bukti wajib diunggah")
		return
	}
	defer file.Close()
	if err := h.service.AddEvidence(id, user.PenggunaID, user.Roles, file, header); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"pesan": "evidence laporan berhasil disimpan"})
}

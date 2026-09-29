package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"sispat/backend/internal/regu"
)

type ReguHandler struct {
	service *regu.Service
}

func NewReguHandler(service *regu.Service) *ReguHandler {
	return &ReguHandler{service: service}
}

type officerRequest struct {
	WargaID uint `json:"warga_id"`
}

func (h *ReguHandler) List(w http.ResponseWriter, r *http.Request) {
	data, err := h.service.List()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "gagal mengambil daftar regu")
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *ReguHandler) Members(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePathID(w, r, "id")
	if !ok {
		return
	}
	data, err := h.service.Members(id)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "gagal mengambil anggota regu")
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *ReguHandler) Officers(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePathID(w, r, "id")
	if !ok {
		return
	}
	data, err := h.service.Officers(id)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "gagal mengambil petugas regu")
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *ReguHandler) AssignDanru(w http.ResponseWriter, r *http.Request) {
	h.assign(w, r, "DANRU")
}

func (h *ReguHandler) AssignWakil(w http.ResponseWriter, r *http.Request) {
	h.assign(w, r, "WAKIL_DANRU")
}

func (h *ReguHandler) assign(w http.ResponseWriter, r *http.Request, jabatan string) {
	groupID, ok := parsePathID(w, r, "id")
	if !ok {
		return
	}
	var request officerRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil || request.WargaID == 0 {
		writeJSONError(w, http.StatusBadRequest, "warga_id tidak valid")
		return
	}
	if err := h.service.AssignOfficer(groupID, request.WargaID, jabatan); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"pesan": jabatan + " berhasil ditetapkan"})
}

func parsePathID(w http.ResponseWriter, r *http.Request, name string) (uint, bool) {
	value, err := strconv.ParseUint(chi.URLParam(r, name), 10, 64)
	if err != nil || value == 0 {
		writeJSONError(w, http.StatusBadRequest, "id tidak valid")
		return 0, false
	}
	return uint(value), true
}

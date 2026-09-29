package httpapi

import (
	"encoding/json"
	"net/http"

	"gorm.io/gorm"

	"sispat/backend/internal/jadwal"
)

type JadwalHandler struct {
	service *jadwal.Service
}

func NewJadwalHandler(service *jadwal.Service) *JadwalHandler {
	return &JadwalHandler{service: service}
}

func (h *JadwalHandler) List(w http.ResponseWriter, r *http.Request) {
	data, err := h.service.List()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "gagal mengambil jadwal ronda")
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *JadwalHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePathID(w, r, "id")
	if !ok {
		return
	}
	data, err := h.service.Get(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			writeJSONError(w, http.StatusNotFound, "jadwal tidak ditemukan")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "gagal mengambil jadwal ronda")
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *JadwalHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input jadwal.CreateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSONError(w, http.StatusBadRequest, "format permintaan tidak valid")
		return
	}
	data, err := h.service.Create(input)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, data)
}

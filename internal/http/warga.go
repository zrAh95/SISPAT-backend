package httpapi

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"sispat/backend/internal/warga"
)

type WargaHandler struct {
	service *warga.Service
}

func NewWargaHandler(service *warga.Service) *WargaHandler {
	return &WargaHandler{service: service}
}

func (h *WargaHandler) List(w http.ResponseWriter, r *http.Request) {
	daftar, err := h.service.List()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "gagal mengambil daftar warga")
		return
	}
	writeJSON(w, http.StatusOK, daftar)
}

func (h *WargaHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	data, err := h.service.Get(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		writeJSONError(w, http.StatusNotFound, "warga tidak ditemukan")
		return
	}
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "gagal mengambil data warga")
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *WargaHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input warga.CreateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSONError(w, http.StatusBadRequest, "format permintaan tidak valid")
		return
	}

	data, err := h.service.Create(input)
	if err != nil {
		if err.Error() == "nama warga sudah digunakan atau nama login bentrok" {
			writeJSONError(w, http.StatusConflict, err.Error())
			return
		}
		if err.Error() == "regu tidak ditemukan" {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, data)
}

func (h *WargaHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	var input warga.UpdateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSONError(w, http.StatusBadRequest, "format permintaan tidak valid")
		return
	}
	data, err := h.service.Update(id, input)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		writeJSONError(w, http.StatusNotFound, "warga tidak ditemukan")
		return
	}
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *WargaHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	if err := h.service.HardDelete(id); err != nil {
		if err.Error() == "warga tidak dapat dihapus karena memiliki riwayat absensi" {
			writeJSONError(w, http.StatusConflict, err.Error())
			return
		}
		log.Printf("gagal hard delete warga id=%d: %v", id, err)
		writeJSONError(w, http.StatusInternalServerError, "gagal menghapus warga")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"pesan": "warga dinonaktifkan"})
}

func (h *WargaHandler) SetStatus(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	var request struct {
		Aktif bool `json:"aktif"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSONError(w, http.StatusBadRequest, "format permintaan tidak valid")
		return
	}
	data, err := h.service.SetActive(id, request.Aktif)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		writeJSONError(w, http.StatusNotFound, "warga tidak ditemukan")
		return
	}
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "gagal mengubah status warga")
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func parseID(w http.ResponseWriter, r *http.Request) (uint, bool) {
	value, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil || value == 0 {
		writeJSONError(w, http.StatusBadRequest, "id warga tidak valid")
		return 0, false
	}
	return uint(value), true
}

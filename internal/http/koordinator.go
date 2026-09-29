package httpapi

import (
	"encoding/json"
	"net/http"

	"sispat/backend/internal/koordinator"
)

type KoordinatorHandler struct {
	service *koordinator.Service
}

func NewKoordinatorHandler(service *koordinator.Service) *KoordinatorHandler {
	return &KoordinatorHandler{service: service}
}

type assignKoordinatorRequest struct {
	WargaID uint `json:"warga_id"`
}

func (h *KoordinatorHandler) ListCalon(w http.ResponseWriter, r *http.Request) {
	data, err := h.service.ListCalon()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "gagal mengambil daftar warga")
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *KoordinatorHandler) Assign(w http.ResponseWriter, r *http.Request) {
	var request assignKoordinatorRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSONError(w, http.StatusBadRequest, "format permintaan tidak valid")
		return
	}
	if err := h.service.Assign(request.WargaID); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"pesan": "koordinator berhasil ditetapkan"})
}

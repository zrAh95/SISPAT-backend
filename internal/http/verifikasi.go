package httpapi

import (
	"encoding/json"
	"net/http"

	"sispat/backend/internal/auth"
	"sispat/backend/internal/verifikasi"
)

type VerifikasiHandler struct{ service *verifikasi.Service }

func NewVerifikasiHandler(service *verifikasi.Service) *VerifikasiHandler {
	return &VerifikasiHandler{service: service}
}

type verificationRequest struct {
	AlasanDitolak string `json:"alasan_ditolak"`
}

func (h *VerifikasiHandler) List(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserFromContext(r.Context())
	id, ok := parsePathID(w, r, "id")
	if !ok {
		return
	}
	data, err := h.service.ListSession(id, user.PenggunaID, user.Roles)
	if err != nil {
		writeJSONError(w, http.StatusForbidden, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *VerifikasiHandler) Verify(w http.ResponseWriter, r *http.Request) {
	h.process(w, r, false)
}

func (h *VerifikasiHandler) Reject(w http.ResponseWriter, r *http.Request) {
	h.process(w, r, true)
}

func (h *VerifikasiHandler) process(w http.ResponseWriter, r *http.Request, reject bool) {
	user, _ := auth.UserFromContext(r.Context())
	id, ok := parsePathID(w, r, "id")
	if !ok {
		return
	}
	var request verificationRequest
	if r.Body != nil && reject {
		_ = json.NewDecoder(r.Body).Decode(&request)
	}
	if err := h.service.Verify(id, user.PenggunaID, user.Roles, reject, request.AlasanDitolak); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"pesan": "verifikasi berhasil"})
}

func (h *VerifikasiHandler) Photo(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserFromContext(r.Context())
	id, ok := parsePathID(w, r, "id")
	if !ok {
		return
	}
	path, mimeType, err := h.service.OpenPhoto(id, user.PenggunaID, user.Roles)
	if err != nil {
		writeJSONError(w, http.StatusForbidden, err.Error())
		return
	}
	w.Header().Set("Content-Type", mimeType)
	http.ServeFile(w, r, path)
}

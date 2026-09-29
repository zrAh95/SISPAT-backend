package httpapi

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"

	"sispat/backend/internal/auth"
)

type Handler struct {
	auth *auth.Service
}

func NewHandler(authService *auth.Service) *Handler {
	return &Handler{auth: authService}
}

type loginRequest struct {
	Username string `json:"nama_pengguna"`
	Password string `json:"kata_sandi"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var request loginRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSONError(w, http.StatusBadRequest, "format permintaan tidak valid")
		return
	}

	user, sessionID, err := h.auth.Login(
		request.Username,
		request.Password,
		clientIP(r),
		r.UserAgent(),
	)
	if err != nil {
		if err.Error() == "terlalu banyak percobaan login" {
			writeJSONError(w, http.StatusTooManyRequests, err.Error())
			return
		}
		if err.Error() == "kredensial tidak valid" {
			writeJSONError(w, http.StatusUnauthorized, err.Error())
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "gagal memproses login")
		return
	}

	h.auth.SetSessionCookie(w, sessionID)
	writeJSON(w, http.StatusOK, user)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if err := h.auth.Logout(h.auth.SessionID(r)); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "gagal logout")
		return
	}

	h.auth.ClearSessionCookie(w)
	writeJSON(w, http.StatusOK, map[string]string{"pesan": "logout berhasil"})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "silakan masuk terlebih dahulu")
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func clientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		return truncateIP(forwarded)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func truncateIP(value string) string {
	if comma := strings.Index(value, ","); comma != -1 {
		value = value[:comma]
	}
	return strings.TrimSpace(value)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"pesan": message})
}

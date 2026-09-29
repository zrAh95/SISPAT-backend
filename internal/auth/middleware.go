package auth

import (
	"crypto/subtle"
	"net/http"
)

const CSRFHeader = "X-CSRF-Token"

func (s *Service) RequireSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := s.CurrentUser(s.SessionID(r))
		if err != nil {
			writeAuthError(w, http.StatusUnauthorized, "silakan masuk terlebih dahulu")
			return
		}
		if isMutation(r.Method) && subtle.ConstantTimeCompare([]byte(r.Header.Get(CSRFHeader)), []byte(user.CSRFToken)) != 1 {
			writeAuthError(w, http.StatusForbidden, "token CSRF tidak valid")
			return
		}
		next.ServeHTTP(w, r.WithContext(WithUser(r.Context(), user)))
	})
}

func isMutation(method string) bool {
	return method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch || method == http.MethodDelete
}

func (s *Service) RequireRoles(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := UserFromContext(r.Context())
			if !ok {
				writeAuthError(w, http.StatusUnauthorized, "silakan masuk terlebih dahulu")
				return
			}
			for _, role := range user.Roles {
				if _, ok := allowed[role]; ok {
					next.ServeHTTP(w, r)
					return
				}
			}
			writeAuthError(w, http.StatusForbidden, "Anda tidak memiliki akses")
		})
	}
}

func writeAuthError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(`{"pesan":"` + message + `"}`))
}

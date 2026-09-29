package httpapi

import (
	"log"
	"net/http"
)

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Printf("panic pada %s %s: %v", r.Method, r.URL.Path, recovered)
				writeJSONError(w, http.StatusInternalServerError, "terjadi kesalahan pada server")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

package middleware

import (
	"encoding/json"
	"log"
	"net/http"
)

type errorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// Recovery es un middleware que recupera de panics con mejor manejo de errores
func Recovery(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("🚨 PANIC RECOVERED: %v | Request: %s %s | Remote: %s", err, r.Method, r.URL.Path, r.RemoteAddr)

				// Asegurar que no se haya enviado respuesta ya
				if w.Header().Get("Content-Type") == "" {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(errorResponse{
						Success: false,
						Error:   "Internal server error occurred",
					})
				} else {
					log.Printf("⚠️  Response already started, cannot send error response")
				}
			}
		}()
		next(w, r)
	}
}

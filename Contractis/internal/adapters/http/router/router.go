package router

import (
	"net/http"

	"github.com/rodascaar/contractis/internal/adapters/http/handlers"
	"github.com/rodascaar/contractis/internal/adapters/http/middleware"
)

// Router configura y retorna el router HTTP
type Router struct {
	uploadHandler   *handlers.UploadHandler
	estimateHandler *handlers.EstimateHandler
	historyHandler  *handlers.HistoryHandler
	staticPath      string
}

// NewRouter crea una nueva instancia de Router
func NewRouter(
	uploadHandler *handlers.UploadHandler,
	estimateHandler *handlers.EstimateHandler,
	historyHandler *handlers.HistoryHandler,
	staticPath string,
) *Router {
	return &Router{
		uploadHandler:   uploadHandler,
		estimateHandler: estimateHandler,
		historyHandler:  historyHandler,
		staticPath:      staticPath,
	}
}

// Setup configura las rutas del servidor
func (r *Router) Setup() *http.ServeMux {
	mux := http.NewServeMux()

	// Health check endpoint (sin middleware para máxima disponibilidad)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","timestamp":"` + http.TimeFormat + `"}`))
	})

	// API endpoints con middleware
	mux.HandleFunc("/upload", r.applyMiddleware(r.uploadHandler.Handle))
	mux.HandleFunc("/estimate", r.applyMiddleware(r.estimateHandler.Handle))

	// History endpoints
	mux.HandleFunc("/api/contracts", r.applyMiddleware(r.historyHandler.HandleList))
	mux.HandleFunc("/api/contracts/search", r.applyMiddleware(r.historyHandler.HandleSearch))
	mux.HandleFunc("/api/contracts/recent", r.applyMiddleware(r.historyHandler.HandleGetRecent))
	mux.HandleFunc("/api/contracts/stats", r.applyMiddleware(r.historyHandler.HandleGetStats))
	mux.HandleFunc("/api/contracts/get", r.applyMiddleware(r.historyHandler.HandleGetByID))
	mux.HandleFunc("/api/contracts/delete", r.applyMiddleware(r.historyHandler.HandleDelete))

	// Archivos estáticos con restricciones de seguridad
	fileServer := http.FileServer(http.Dir(r.staticPath))
	mux.Handle("/", r.secureStaticFileServer(fileServer))

	return mux
}

// applyMiddleware aplica los middlewares a un handler
func (r *Router) applyMiddleware(handler http.HandlerFunc) http.HandlerFunc {
	return middleware.Recovery(middleware.Logging(handler))
}

// secureStaticFileServer añade headers de seguridad a los archivos estáticos
func (r *Router) secureStaticFileServer(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Headers de seguridad para archivos estáticos
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Cache control para archivos estáticos
		if req.URL.Path != "/" {
			w.Header().Set("Cache-Control", "public, max-age=31536000") // 1 año
		}

		handler.ServeHTTP(w, req)
	})
}

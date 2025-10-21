package http

import (
	"fmt"
	"net/http"

	"github.com/rodascaar/contractis/internal/adapters/http/router"
)

// Server representa el servidor HTTP
type Server struct {
	port   string
	router *router.Router
}

// NewServer crea una nueva instancia de Server
func NewServer(port string, router *router.Router) *Server {
	return &Server{
		port:   port,
		router: router,
	}
}

// Start inicia el servidor HTTP
func (s *Server) Start() error {
	mux := s.router.Setup()

	fmt.Printf("🚀 Contractis servidor iniciado en http://localhost:%s\n", s.port)
	return http.ListenAndServe(":"+s.port, mux)
}

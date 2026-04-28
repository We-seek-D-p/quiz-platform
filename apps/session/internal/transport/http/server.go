package httptransport

import (
	"context"
	"net/http"
	"time"
)

type Server struct {
	server *http.Server
}

func NewServer(addr string, handler http.Handler) *Server {
	return &Server{
		server: &http.Server{
			Addr:              addr,
			Handler:           handler,
			ReadHeaderTimeout: 5 * time.Second,
		},
	}
}

func (s *Server) Run() error {
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

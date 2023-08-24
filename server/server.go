package server

import (
	"context"
	"crypto/rand"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/csrf"
)

type Server struct {
	srv *http.Server
}

type RegisterRoutes interface {
	RegisterRoutes(mux *chi.Mux)
}

func NewServer(addr string, routes ...RegisterRoutes) (*Server, error) {
	mux := chi.NewMux()

	for _, r := range routes {
		r.RegisterRoutes(mux)
	}

	csrfSecret, err := genCSRFSecret()
	if err != nil {
		return nil, err
	}
	csrfProtect := csrf.Protect(csrfSecret, csrf.CookieName("stuff.csrf.token"), csrf.FieldName("stuff.csrf.token"))

	handler := csrfProtect(mux)

	return &Server{
		srv: &http.Server{
			Addr:    addr,
			Handler: handler,
		},
	}, nil
}

func (s *Server) Start(ctx context.Context) error {
	slog.InfoContext(ctx, "starting http server on "+s.srv.Addr)
	if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil

}

func (s *Server) Stop(ctx context.Context) error {
	slog.InfoContext(ctx, "shutting down http server")
	return s.srv.Shutdown(ctx)
}

func genCSRFSecret() ([]byte, error) {
	var b [32]byte

	_, err := rand.Read(b[:])
	if err != nil {
		return nil, err
	}

	return b[:], nil
}

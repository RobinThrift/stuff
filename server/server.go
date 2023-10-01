package server

import (
	"context"
	"crypto/rand"
	"log/slog"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/kodeshack/stuff/static"
)

type Server struct {
	srv *http.Server
}

type RegisterRoutes func(mux *chi.Mux)

func NewServer(addr string, sm *scs.SessionManager, routes ...RegisterRoutes) (*Server, error) {
	srv := &Server{}

	mux := chi.NewMux()

	csrfMiddleware, err := csrfMiddleware([]string{"/static"})
	if err != nil {
		return nil, err
	}

	mux.Use(
		requestIDMiddleware,
		logReqMiddleware,
		sessionMiddleware(sm),
		csrfMiddleware,
		loginRedirectMiddleware([]string{"/login", "/auth/changepassword", "/static/"}),
	)

	mux.Get("/health", http.HandlerFunc(srv.handleHealth))
	mux.Handle("/static/*", static.Files("/static/"))

	for _, r := range routes {
		r(mux)
	}

	return &Server{
		srv: &http.Server{
			Addr:    addr,
			Handler: mux,
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

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("ok"))
	if err != nil {
		slog.ErrorContext(r.Context(), "error writing http health response", "error", err)
	}
}

func genCSRFSecret() ([]byte, error) {
	var b [32]byte

	_, err := rand.Read(b[:])
	if err != nil {
		return nil, err
	}

	return b[:], nil
}

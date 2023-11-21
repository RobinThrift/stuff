package server

import (
	"context"
	"crypto/rand"
	"log/slog"
	"net/http"

	"github.com/RobinThrift/stuff/frontend"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
)

type Server struct {
	Mux *chi.Mux
	srv *http.Server
}

func NewServer(addr string, useSecureCookies bool, sm *scs.SessionManager) (*Server, error) {
	srv := &Server{}

	mux := chi.NewMux()

	csrfMiddleware, err := csrfMiddleware(useSecureCookies, []string{"/static"})
	if err != nil {
		return nil, err
	}

	mux.Use(
		requestIDMiddleware,
		logReqMiddleware,
		sessionMiddleware(sm, []string{"/static", "/manifest"}),
		csrfMiddleware,
		loginRedirectMiddleware([]string{"/login", "/auth/changepassword", "/static/", "/manifest/"}),
	)

	mux.Get("/health", http.HandlerFunc(srv.handleHealth))
	mux.Handle("/static/*", frontend.Files("/static/"))
	mux.Handle("/manifest/*", frontend.Manifest())

	return &Server{
		Mux: mux,
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
	slog.InfoContext(ctx, "stopping http server")
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

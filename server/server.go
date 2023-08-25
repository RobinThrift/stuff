package server

import (
	"context"
	"crypto/rand"
	"log/slog"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/csrf"
	"github.com/kodeshack/stuff"
)

type Server struct {
	srv *http.Server
}

type RegisterRoutes func(mux *chi.Mux)

func NewServer(addr string, routes ...RegisterRoutes) (*Server, error) {
	srv := &Server{}

	sessionManager := scs.New()
	sessionManager.Lifetime = 24 * time.Hour
	sessionManager.Cookie.HttpOnly = true
	sessionManager.Cookie.Persist = true
	sessionManager.Cookie.SameSite = http.SameSiteStrictMode

	mux := chi.NewMux()

	mux.Use(
		requestIDMiddleware,
		logReqMiddleware,
		sessionMiddleware(sessionManager),
		loginRedirectMiddleware([]string{"/login", "/auth/changepassword", "/static/"}),
	)

	mux.Get("/health", http.HandlerFunc(srv.handleHealth))
	mux.Handle("/static/*", stuff.StaticFiles("/static/"))

	for _, r := range routes {
		r(mux)
	}

	csrfSecret, err := genCSRFSecret()
	if err != nil {
		return nil, err
	}
	csrfProtect := csrf.Protect(
		csrfSecret,
		csrf.CookieName("stuff.csrf.token"),
		csrf.FieldName("stuff.csrf.token"),
		csrf.ErrorHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/login", http.StatusFound)
		})),
	)

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

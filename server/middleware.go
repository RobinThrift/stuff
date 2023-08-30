package server

import (
	"encoding/gob"
	"log/slog"
	"net/http"
	"strings"

	"github.com/alexedwards/scs/v2"
	"github.com/gorilla/csrf"
	"github.com/kodeshack/stuff/ctxx"
	"github.com/kodeshack/stuff/server/session"
	"github.com/kodeshack/stuff/users"
	"github.com/segmentio/ksuid"
)

func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := ctxx.RequestIDWithCtx(r.Context(), ksuid.New().String())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func logReqMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.InfoContext(r.Context(), r.URL.Path, "method", r.Method)
		next.ServeHTTP(w, r)
	})
}

func sessionMiddleware(sessionManager *scs.SessionManager) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := session.CtxWithSessionManager(r.Context(), sessionManager)
			sessionManager.LoadAndSave(next).ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func loginRedirectMiddleware(skipFor []string) func(next http.Handler) http.Handler {
	gob.Register(&users.User{})

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, s := range skipFor {
				if strings.HasPrefix(r.URL.Path, s) {
					next.ServeHTTP(w, r)
					return
				}
			}

			_, ok := session.Get[any](r.Context(), "user")
			if ok {
				next.ServeHTTP(w, r)
				return
			}

			http.Redirect(w, r, "/login", http.StatusFound)
		})
	}
}

func csrfMiddleware() (func(next http.Handler) http.Handler, error) {
	csrfSecret, err := genCSRFSecret()
	if err != nil {
		return nil, err
	}

	csrfProtect := csrf.Protect(
		csrfSecret,
		csrf.CookieName("stuff.csrf.token"),
		csrf.FieldName("stuff.csrf.token"),
		csrf.ErrorHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			slog.ErrorContext(r.Context(), "csrf error handler", "error", csrf.FailureReason(r))
			session.Put(r.Context(), "csrf_error", csrf.FailureReason(r).Error())
			http.Redirect(w, r, r.URL.String(), http.StatusFound)
		})),
	)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			csrfProtect(next).ServeHTTP(w, r)
		})
	}, nil

}

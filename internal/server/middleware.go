package server

import (
	"encoding/gob"
	"log/slog"
	"net/http"
	"strings"

	"github.com/RobinThrift/stuff/auth"
	"github.com/RobinThrift/stuff/internal/requestid"
	"github.com/RobinThrift/stuff/internal/server/session"
	"github.com/RobinThrift/stuff/views"
	"github.com/alexedwards/scs/v2"
	"github.com/gorilla/csrf"
	"github.com/segmentio/ksuid"
)

func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := requestid.WithCtx(r.Context(), ksuid.New().String())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func logReqMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.Path
		if r.URL.RawQuery != "" {
			url += "?" + r.URL.RawQuery
		}
		slog.InfoContext(r.Context(), url, "method", r.Method)
		next.ServeHTTP(w, r)
	})
}

func sessionMiddleware(sessionManager *scs.SessionManager, skipFor []string) func(next http.Handler) http.Handler {
	gob.Register(&auth.User{})
	gob.Register(views.FlashMessage{})

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, s := range skipFor {
				if strings.HasPrefix(r.URL.Path, s) {
					next.ServeHTTP(w, r)
					return
				}
			}

			ctx := session.CtxWithSessionManager(r.Context(), sessionManager)
			sessionManager.LoadAndSave(next).ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func loginRedirectMiddleware(skipFor []string) func(next http.Handler) http.Handler {
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

func csrfMiddleware(secure bool, skipFor []string) (func(next http.Handler) http.Handler, error) {
	csrfSecret, err := genCSRFSecret()
	if err != nil {
		return nil, err
	}

	csrfProtect := csrf.Protect(
		csrfSecret,
		csrf.Path("/"),
		csrf.Secure(secure),
		csrf.SameSite(csrf.SameSiteStrictMode),
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
			for _, s := range skipFor {
				if strings.HasPrefix(r.URL.Path, s) {
					next.ServeHTTP(w, r)
					return
				}
			}

			csrfProtect(next).ServeHTTP(w, r)
		})
	}, nil

}

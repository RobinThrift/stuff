package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
)

// Error Follows RFC7807 (https://datatracker.ietf.org/doc/html/rfc7807)
type Error struct {
	Code   int    `json:"code"`
	Title  string `json:"title"`
	Detail string `json:"detail"`
	Type   string `json:"type"`
}

func (e Error) Error() string {
	var b strings.Builder
	b.WriteString(e.Title)

	if e.Code != 0 {
		b.WriteString(" (")
		b.WriteString(fmt.Sprint(e.Code))
		b.WriteString(")")
	}

	if e.Detail != "" {
		b.WriteString(": ")
		b.WriteString(e.Detail)
	}

	return b.String()
}

func RespondWithError(ctx context.Context, w http.ResponseWriter, err error) {
	var apiErr Error
	if !errors.As(err, &apiErr) {
		apiErr.Code = http.StatusInternalServerError
		apiErr.Title = err.Error()
		apiErr.Title = "stuff/internal-server-error"
	}

	b, err := json.Marshal(apiErr)
	if err != nil {
		slog.ErrorContext(ctx, "error marshalling api error JSON", "error", err)
		return
	}

	AddJSONContentType(w)

	_, err = w.Write(b)
	if err != nil {
		slog.ErrorContext(ctx, "error writing to HTTP response", "error", err)
	}
}

func AddJSONContentType(w http.ResponseWriter) {
	w.Header().Add("content-type", "application/json")
}

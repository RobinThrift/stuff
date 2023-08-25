package ctxx

import (
	"context"
)

type ctxReqIDKeyType string

const ctxReqIDKey = ctxReqIDKeyType("ctxReqIDKey")

func RequestIDWithCtx(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxReqIDKey, id)
}

func RequestIDFromCtx(ctx context.Context) (string, bool) {
	val := ctx.Value(ctxReqIDKey)
	id, ok := val.(string)
	if ok {
		return id, true
	}

	return "", false
}

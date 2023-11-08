package requestid

import (
	"context"
)

type ctxReqIDKeyType string

const ctxReqIDKey = ctxReqIDKeyType("ctxReqIDKey")

func WithCtx(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxReqIDKey, id)
}

func FromCtx(ctx context.Context) (string, bool) {
	val := ctx.Value(ctxReqIDKey)
	id, ok := val.(string)
	if ok {
		return id, true
	}

	return "", false
}

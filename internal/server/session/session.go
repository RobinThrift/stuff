package session

import (
	"context"

	"github.com/alexedwards/scs/v2"
)

type ctxSessMngrKeyType string

const ctxSessMngrKey = ctxSessMngrKeyType("ctxSessionKey")

func CtxWithSessionManager(ctx context.Context, sm *scs.SessionManager) context.Context {
	return context.WithValue(ctx, ctxSessMngrKey, sm)
}

func RenewToken(ctx context.Context) error {
	sm, ok := ctx.Value(ctxSessMngrKey).(*scs.SessionManager)
	if !ok {
		return nil
	}

	return sm.RenewToken(ctx)
}

func Put(ctx context.Context, key string, value any) {
	sm, ok := ctx.Value(ctxSessMngrKey).(*scs.SessionManager)
	if !ok {
		return
	}

	sm.Put(ctx, key, value)
}

func Get[V any](ctx context.Context, key string) (V, bool) {
	var defaultVal V

	sm, ok := ctx.Value(ctxSessMngrKey).(*scs.SessionManager)
	if !ok {
		return defaultVal, false
	}

	val, ok := sm.Get(ctx, key).(V)
	if !ok {
		return defaultVal, false
	}

	return val, true
}

func Pop[V any](ctx context.Context, key string) (V, bool) {
	var defaultVal V

	sm, ok := ctx.Value(ctxSessMngrKey).(*scs.SessionManager)
	if !ok {
		return defaultVal, false
	}

	val, ok := sm.Pop(ctx, key).(V)
	if !ok {
		return defaultVal, false
	}

	return val, true
}

func Remove(ctx context.Context, key string) {
	sm, ok := ctx.Value(ctxSessMngrKey).(*scs.SessionManager)
	if !ok {
		return
	}

	sm.Remove(ctx, key)
}

func Destroy(ctx context.Context) error {
	sm, ok := ctx.Value(ctxSessMngrKey).(*scs.SessionManager)
	if !ok {
		return nil
	}

	return sm.Destroy(ctx)
}

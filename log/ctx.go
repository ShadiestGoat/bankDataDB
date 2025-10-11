package log

import (
	"context"
)

type CtxLogger func(context.Context) Logger

func NewCtxLogger() CtxLogger {
	return NewCtxLoggerComplex(GlobalLogger)
}

func NewCtxLoggerComplex(fallback Logger) CtxLogger {
	return func(ctx context.Context) Logger {
		if l := ctx.Value(ctxKey); l != nil {
			return l.(Logger)
		}

		return fallback
	}
}

func (b CtxLogger) With(kv ...any) CtxLogger {
	return func(ctx context.Context) Logger {
		return b(ctx).With(kv...)
	}
}

type ctxKeyT int

const ctxKey ctxKeyT = 0

func ContextSet(ctx context.Context, l Logger, kv ...any) context.Context {
	return context.WithValue(ctx, ctxKey, l.With(kv...))
}

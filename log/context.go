package log

import (
	"context"
)

type key int

const (
	logContextKey key = iota
)

// WithContext 返回设置日志值的上下文副本.
func WithContext(ctx context.Context) context.Context {
	return std.WithContext(ctx)
}

func (l *zapLogger) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, logContextKey, l)
}

// FromContext 返回 ctx 上日志键的值.
func FromContext(ctx context.Context) Logger {
	if ctx != nil {
		logger := ctx.Value(logContextKey)
		if logger != nil {
			return logger.(Logger)
		}
	}

	return WithName("Unknown-Context")
}

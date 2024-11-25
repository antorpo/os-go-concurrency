package log

import (
	"context"

	"go.uber.org/zap"
)

type logCtxKey struct{}

func Context(ctx context.Context, log Logger) context.Context {
	l, ok := log.(*logger)
	if ok {
		l.Logger.WithOptions(zap.AddCallerSkip(1))
	}
	return context.WithValue(ctx, logCtxKey{}, log)
}

func FromContext(ctx context.Context) Logger {
	l, _ := ctx.Value(logCtxKey{}).(Logger)
	return l
}

func Sugar(ctx context.Context) *SugaredLogger {
	return getLogger(ctx).Sugar()
}

func Named(ctx context.Context, s string) context.Context {
	logger := getLogger(ctx).Named(s)
	return context.WithValue(ctx, logCtxKey{}, logger)
}

func With(ctx context.Context, fields ...Field) context.Context {
	logger := getLogger(ctx).With(fields...)
	return context.WithValue(ctx, logCtxKey{}, logger)
}

func WithLevel(ctx context.Context, lvl Level) context.Context {
	logger := getLogger(ctx).WithLevel(lvl)
	return context.WithValue(ctx, logCtxKey{}, logger)
}

func Check(ctx context.Context, lvl Level, msg string) *CheckedEntry {
	return getLogger(ctx).Check(lvl, msg)
}

func DPanic(ctx context.Context, msg string, fields ...Field) {
	getLogger(ctx).DPanic(msg, fields...)
}

func Debug(ctx context.Context, msg string, fields ...Field) {
	getLogger(ctx).Debug(msg, fields...)
}

func Error(ctx context.Context, msg string, fields ...Field) {
	getLogger(ctx).Error(msg, fields...)
}

func Fatal(ctx context.Context, msg string, fields ...Field) {
	getLogger(ctx).Fatal(msg, fields...)
}

func Info(ctx context.Context, msg string, fields ...Field) {
	getLogger(ctx).Info(msg, fields...)
}

func Panic(ctx context.Context, msg string, fields ...Field) {
	getLogger(ctx).Panic(msg, fields...)
}

func Warn(ctx context.Context, msg string, fields ...Field) {
	getLogger(ctx).Warn(msg, fields...)
}

func getLogger(ctx context.Context) Logger {
	l, ok := ctx.Value(logCtxKey{}).(Logger)
	if ok {
		return l
	}
	return DefaultLogger
}

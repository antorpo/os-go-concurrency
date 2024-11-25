package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type CheckedEntry = zapcore.CheckedEntry

type SugaredLogger = zap.SugaredLogger

type Logger interface {
	Check(lvl Level, msg string) *CheckedEntry

	Named(s string) Logger

	Sugar() *SugaredLogger

	With(fields ...Field) Logger

	WithLevel(lvl Level) Logger

	DPanic(msg string, fields ...Field)

	Debug(msg string, fields ...Field)

	Error(msg string, fields ...Field)

	Fatal(msg string, fields ...Field)

	Info(msg string, fields ...Field)

	Panic(msg string, fields ...Field)

	Warn(msg string, fields ...Field)

	Level() Level
}

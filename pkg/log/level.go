package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Level = zapcore.Level

const (
	DebugLevel = zapcore.DebugLevel

	InfoLevel = zapcore.InfoLevel

	WarnLevel = zapcore.WarnLevel

	ErrorLevel = zapcore.ErrorLevel

	DPanicLevel = zapcore.DPanicLevel

	PanicLevel = zapcore.PanicLevel

	FatalLevel = zapcore.FatalLevel
)

type AtomicLevel = zap.AtomicLevel

func NewAtomicLevel() AtomicLevel {
	return zap.NewAtomicLevel()
}

func NewAtomicLevelAt(l Level) AtomicLevel {
	a := NewAtomicLevel()
	a.SetLevel(l)
	return a
}

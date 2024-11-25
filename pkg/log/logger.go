package log

import (
	"io"
	"os"
	"time"

	"github.com/antorpo/os-go-concurrency/pkg/log/encoders"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var DefaultLogger Logger = &logger{
	Logger: zap.NewNop(),
}

func NewProductionLogger(lvl *AtomicLevel, opts ...Option) Logger {
	opts = append(_defaultOption, opts...)

	var cfg logConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	var zapOptions []zap.Option

	if cfg.caller {
		zapOptions = append(zapOptions, zap.AddCaller(), zap.AddCallerSkip(cfg.callerSkip))
	}

	if cfg.stacktrace {
		zapOptions = append(zapOptions, zap.AddStacktrace(zap.ErrorLevel))
	}

	zapOptions = append(zapOptions, wrapCoreWithLevel(lvl))

	l := zap.New(newZapCoreAtLevel(zap.DebugLevel, cfg), zapOptions...)

	return &logger{
		Logger: l,
	}
}

type logger struct {
	*zap.Logger
}

var _ Logger = (*logger)(nil)

func (l *logger) WithLevel(level Level) Logger {
	lvl := zap.NewAtomicLevelAt(level)
	child := l.Logger.WithOptions(wrapCoreWithLevel(&lvl))
	return &logger{
		Logger: child,
	}
}

func (l *logger) With(fields ...Field) Logger {
	child := l.Logger.With(fields...)
	return &logger{
		Logger: child,
	}
}

func (l *logger) Named(s string) Logger {
	child := l.Logger.Named(s)
	return &logger{
		Logger: child,
	}
}

func (l *logger) Level() Level {
	return zapcore.LevelOf(l.Core())
}

type WriteSyncer interface {
	io.Writer
	Sync() error
}

type encoderFactory func(config zapcore.EncoderConfig) zapcore.Encoder

type logConfig struct {
	levelKey   string
	caller     bool
	callerSkip int
	stacktrace bool
	writer     WriteSyncer

	encoderFactory encoderFactory
}

type Option func(s *logConfig)

func WithLevelKey(key string) Option {
	return func(s *logConfig) {
		s.levelKey = key
	}
}

func WithCaller(t bool) Option {
	return func(s *logConfig) {
		s.caller = t
	}
}

func WithCallerSkip(skip int) Option {
	return func(s *logConfig) {
		s.callerSkip = skip
	}
}

func WithStacktraceOnError(b bool) Option {
	return func(s *logConfig) {
		s.stacktrace = b
	}
}

func WithJSONEncoding() Option {
	return func(s *logConfig) {
		s.encoderFactory = func(config zapcore.EncoderConfig) zapcore.Encoder {
			return zapcore.NewJSONEncoder(config)
		}
	}
}

func WithConsoleEncoding() Option {
	return func(s *logConfig) {
		s.encoderFactory = func(config zapcore.EncoderConfig) zapcore.Encoder {
			return zapcore.NewConsoleEncoder(config)
		}
	}
}

func WithKeyValueEncoding(kveOption ...encoders.KeyValueEncoderOption) Option {
	return func(s *logConfig) {
		s.encoderFactory = func(config zapcore.EncoderConfig) zapcore.Encoder {
			return encoders.NewKeyValueEncoder(config, kveOption...)
		}
	}
}

func WithWriter(w WriteSyncer) Option {
	return func(s *logConfig) {
		s.writer = w
	}
}

var (
	_stderr        = zapcore.Lock(zapcore.AddSync(os.Stderr))
	_defaultOption = []Option{
		WithWriter(_stderr),
		WithLevelKey("level"),
		WithStacktraceOnError(true),
		WithCaller(true),
		WithCallerSkip(1),
		WithKeyValueEncoding(),
	}
)

func newZapCoreAtLevel(lvl zapcore.Level, cfg logConfig) zapcore.Core {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       cfg.levelKey,
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     rfc3399NanoTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	return zapcore.NewCore(cfg.encoderFactory(encoderConfig), cfg.writer, lvl)
}

func rfc3399NanoTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	const RFC3339Micro = "2006-01-02T15:04:05.000000Z07:00"

	enc.AppendString(t.UTC().Format(RFC3339Micro))
}

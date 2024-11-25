package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type coreWithLevel struct {
	zapcore.Core

	lvl *zap.AtomicLevel
}

func (c *coreWithLevel) Enabled(level zapcore.Level) bool {
	return c.lvl.Enabled(level) && c.Core.Enabled(level)
}

func (c *coreWithLevel) Check(e zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if !c.lvl.Enabled(e.Level) {
		return ce
	}
	return c.Core.Check(e, ce)
}

func (c *coreWithLevel) With(fields []zapcore.Field) zapcore.Core {
	core := c.Core.With(fields)
	return &coreWithLevel{
		Core: core,
		lvl:  c.lvl,
	}
}

func wrapCoreWithLevel(l *zap.AtomicLevel) zap.Option {
	return zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		newCore := &coreWithLevel{
			Core: core,
			lvl:  l,
		}

		lvlCore, ok := core.(*coreWithLevel)
		if ok {
			newCore.Core = lvlCore.Core
		}

		return newCore
	})
}

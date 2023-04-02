package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// coreWithLevel is a struct that wraps around a zapcore.Core and a zapcore.Level
// It implements the Enabled and Check methods required by the zapcore.Core interface
// to check whether a log level is enabled and to add the core to a checked entry, respectively.
type coreWithLevel struct {
	zapcore.Core
	level zapcore.Level
}

// Enabled returns true if the provided level is enabled for logging by the core.
// It calls the Enabled method of the wrapped zapcore.Level.
func (c *coreWithLevel) Enabled(l zapcore.Level) bool {
	return c.level.Enabled(l)
}

// Check adds the core to a checked entry if the provided entry's level is enabled for logging.
// It returns the checked entry with the core added or the original checked entry if the level is disabled.
func (c *coreWithLevel) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}

	return ce
}

// With returns a new core with the given fields added to the wrapped core.
// It returns a new coreWithLevel with the same level as the original core.
func (c *coreWithLevel) With(fields []zapcore.Field) zapcore.Core {
	return &coreWithLevel{
		c.Core.With(fields),
		c.level,
	}
}

// WithLevel is an option that creates a logger with the specified log level from an existing logger.
// It returns a zap.Option that wraps the existing core with a coreWithLevel with the specified level.
func WithLevel(lvl zapcore.Level) zap.Option {
	return zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return &coreWithLevel{core, lvl}
	})
}

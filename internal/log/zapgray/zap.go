package zapgray

import (
	"github.com/liqifyl/chat-go/internal/log/gelf"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// map zapcore's log levels to standard syslog levels used by gelf, approximately.
var zapToSyslog = map[zapcore.Level]uint{
	zapcore.DebugLevel:  7,
	zapcore.InfoLevel:   6,
	zapcore.WarnLevel:   4,
	zapcore.ErrorLevel:  3,
	zapcore.DPanicLevel: 2,
	zapcore.PanicLevel:  2,
	zapcore.FatalLevel:  1,
}

func SyslogLevelEncoder(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendUint(zapToSyslog[l])
}

func NewGraylogEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "_logger",
		CallerKey:      "_caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "short_message",
		StacktraceKey:  "_stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    SyslogLevelEncoder,
		EncodeTime:     zapcore.EpochTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

func ZapGrayCore(addr string, lv zap.AtomicLevel) zapcore.Core {
	return NewGelfCore(gelf.New(gelf.Config{
		GraylogAddr: addr,
	}), lv)
}


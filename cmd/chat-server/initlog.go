package main

import (
	"fmt"
	"github.com/liqifyl/chat-go/internal/log/zapgray"
	"os"
	"strconv"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/alecthomas/kingpin.v2"
)

func ZapConsoleCore() (core zapcore.Core) {

	core = zapcore.NewNopCore()
	if !LogToStdout {
		return
	}

	lv := zapcore.FatalLevel
	if err := lv.UnmarshalText([]byte(LogLevelStdout)); err != nil {
		return
	}

	consoleWriter := zapcore.Lock(os.Stdout)

	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)

	alv := zap.NewAtomicLevelAt(lv)

	core = zapcore.NewCore(consoleEncoder, consoleWriter, alv)
	return
}

func ZapGraylogCore() (core zapcore.Core) {
	core = zapcore.NewNopCore()
	if !LogToGraylog {
		return
	}
	lv := zapcore.FatalLevel
	if err := lv.UnmarshalText([]byte(LogLevelGraylog)); err != nil {
		return
	}
	host := (LogToGraylogAddress).Hostname()
	port, _ := strconv.Atoi((LogToGraylogAddress).Port())
	fmt.Printf("ZapGraylogCore->host:%v,%d", host, port)

	if host == "" || port <= 0 {
		return
	}

	core = zapgray.ZapGrayCore(fmt.Sprintf("%s:%d", host, port), zap.NewAtomicLevelAt(lv))
	return
}

func InitLog() {
	cc := ZapConsoleCore()
	gc := ZapGraylogCore()
	opts := []zap.Option{zap.AddCaller(), zap.AddStacktrace(zapcore.DPanicLevel)}
	if LogDevelopment {
		opts = append(opts, zap.Development())
	}
	l := zap.New(zapcore.NewTee(cc, gc), opts...).Named(kingpin.CommandLine.Name)
	zap.ReplaceGlobals(l)
}

func TermLog() {
	_ = zap.L().Sync()
}

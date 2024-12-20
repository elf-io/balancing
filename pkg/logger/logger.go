// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0

package logger

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"strings"
)

func NewStdoutLogger(loglevel string, Name string) *zap.Logger {

	// log level
	l := zapcore.DebugLevel
	switch strings.ToLower(loglevel) {
	case "debug":
		l = zapcore.DebugLevel
	case "info":
		l = zapcore.InfoLevel
	case "warn":
		l = zapcore.WarnLevel
	case "error":
		l = zapcore.ErrorLevel
	case "fatal":
		l = zapcore.FatalLevel
	case "panic":
		l = zapcore.PanicLevel
	case "":
		l = zapcore.DebugLevel
	default:
		panic(fmt.Sprintf("unknown log level %s", loglevel))
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = nil
	encoderConfig.EncodeCaller = nil
	encoderConfig.EncodeLevel = nil
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.NameKey = Name

	m := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		zap.NewAtomicLevelAt(l),
	)
	// configures the Logger to annotate each message with the filename, line number, and function name of zap's caller
	t := zap.AddCaller()
	logger := zap.New(m, t)

	return logger
}

package app

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.SugaredLogger

func init() {
	cfg := zap.NewProductionConfig()
	ecfg := zap.NewProductionEncoderConfig()
	ecfg.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncoderConfig = ecfg
	l, _ := cfg.Build()
	defer l.Sync() // flushes buffer, if any
	sugar := l.Sugar()
	logger = sugar
}

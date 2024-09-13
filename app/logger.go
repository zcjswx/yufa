package app

import (
	"fmt"

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
	defer func() {
		if err := l.Sync(); err != nil {
			fmt.Printf("Failed to sync logger: %v\n", err)
		}
	}()
	sugar := l.Sugar()
	logger = sugar
}

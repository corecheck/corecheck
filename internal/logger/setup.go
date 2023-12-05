package logger

import (
	"github.com/corecheck/corecheck/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New() *zap.SugaredLogger {
	var logger *zap.Logger

	if config.Config.Dev {
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		logger, _ = config.Build()
	} else {
		logger, _ = zap.NewProduction()
	}

	return logger.Sugar()
}

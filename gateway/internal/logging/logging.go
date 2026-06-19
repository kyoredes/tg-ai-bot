package logging

import (
	"errors"

	"go.uber.org/zap"
)

var Logger *zap.Logger

func InitLogger(mode string) error {
	var cfg zap.Config
	var err error

	switch mode {
	case "json":
		cfg = zap.NewProductionConfig()
	case "text":
		cfg = zap.NewDevelopmentConfig()
		cfg.Encoding = "console"
	default:
		return errors.New("invalid log format")
	}
	Logger, err = cfg.Build()
	if err != nil {
		return err
	}
	return nil
}

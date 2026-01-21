package logger

import "go.uber.org/zap"

func Newlogger() *zap.Logger {
	log,_ := zap.NewProduction()
	return log
}
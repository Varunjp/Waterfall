package logger

import "go.uber.org/zap"

func NewLogger() *zap.Logger {
	log,_ := zap.NewProduction()
	return log 
}
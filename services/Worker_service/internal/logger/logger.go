package logger

import "go.uber.org/zap"

var log *zap.Logger

func Init() {
	l,_ := zap.NewProduction()
	log = l 
}

func NewLog()*zap.Logger {
	return log 
}
package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func Newlogger(serviceName string) (*zap.Logger,error) {

	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename: "logs/app.log",
		MaxSize: 100,
		MaxBackups: 5,
		MaxAge: 30,
		Compress: true,
	})

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	jsonEncoder := zapcore.NewJSONEncoder(encoderConfig)
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)

	level := zap.InfoLevel

	core := zapcore.NewTee(
		zapcore.NewCore(jsonEncoder,fileWriter,level),
		zapcore.NewCore(consoleEncoder,zapcore.AddSync(os.Stdout),level),
	)
	
	logger := zap.New(core,
		zap.AddCaller(),
		zap.AddStacktrace(zap.ErrorLevel),
		zap.Fields(
			zap.String("service",serviceName),
		),
	)
	
	return logger,nil 
}
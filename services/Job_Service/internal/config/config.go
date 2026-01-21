package config

import (
	"os"
)

type Config struct {
	ServiceName string
	KafkaBrokers []string
	KafkaTopic	string 
	JWTKey		string 
	PORT 		string 
}

func Load() *Config {
	return &Config{
		ServiceName: "job-service",
		KafkaBrokers: []string{os.Getenv("KAFKA_BROKER")},
		KafkaTopic: os.Getenv("KAFKA_JOB_TOPIC"),
		JWTKey: os.Getenv("JWT_KEY"),
		PORT: os.Getenv("JOB_SERVICE_PORT"),
	}
}

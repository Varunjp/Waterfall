package config

import (
	"log"
	"os"
)

type Config struct {
	ServiceName string
	KafkaBrokers []string
	KafkaTopic	string 
	DBDSN        string
	JWTKey		string 
	PORT 		string 
	Topic 		 string 
}

func Load() *Config {
	return &Config{
		ServiceName: "job-service",
		KafkaBrokers: []string{os.Getenv("KAFKA_BROKER")},
		KafkaTopic: os.Getenv("KAFKA_JOB_TOPIC"),
		JWTKey: os.Getenv("JWT_KEY"),
		PORT: os.Getenv("JOB_SERVICE_PORT"),
		DBDSN: must("DB_DSN"),
		Topic: must("QUEUE_TOPIC"),
	}
}

func must(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("missing env var %s", key)
	}
	return v
}
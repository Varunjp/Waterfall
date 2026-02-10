package config

import (
	"log"
	"os"
)

type Config struct {
	ServiceAddr  string
	DBDSN        string
	JWTSecret    string
	JobQueueAddr string
	Topic 		 string 
}

func Load() *Config {
	cfg := &Config{
		ServiceAddr:  get("SERVICE_ADDR", ":50053"),
		DBDSN:        must("DB_DSN"),
		JWTSecret:    must("JWT_SECRET"),
		JobQueueAddr: must("JOB_QUEUE_ADDR"),
		Topic: must("QUEUE_TOPIC"),
	}
	return cfg
}

func get(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func must(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("missing env var %s", key)
	}
	return v
}
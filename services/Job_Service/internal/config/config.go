package config

import (
	"os"
	"strings"
)

type Config struct {
	ServiceName string
	GRPCPort    string

	KafkaBrokers []string
	RedisAddr    string

	OtelEndpoint string
}

func Load() *Config {
	return &Config{
		ServiceName: "job-service",
		GRPCPort:    getEnv("GRPC_PORT", "50051"),

		KafkaBrokers: strings.Split(getEnv("KAFKA_BROKERS","localhost:9092"),","),
		RedisAddr: getEnv("REDIS_ADDR","localhost:6379"),

		OtelEndpoint: getEnv("OTEL_EXPORTER_OTLP_ENDPOINT","localhost:4317"),
	}
}

func getEnv(k,d string) string {
	if v := os.Getenv(k); v != "" {
		return v 
	}
	return d 
}
package config

import "os"

type Config struct {
	ServiceName string 
	GRPCPort 	string 
	RedisURL    string 
	KafkaBroker string
	KafkaTopic  string 
	AdminDBURL  string 
}

func Load() *Config {
	return &Config{
		ServiceName: os.Getenv("SERVICE_NAME"),
		GRPCPort: os.Getenv("GRPC_PORT"),
		RedisURL: os.Getenv("REDIS_URL"),
		KafkaBroker: os.Getenv("KAFKA_BROKER"),
		KafkaTopic: os.Getenv("KAFKA_TOPIC"),
		AdminDBURL: os.Getenv("ADMIN_DB_URL"),
	}
}
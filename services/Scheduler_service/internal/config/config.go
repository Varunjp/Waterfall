package config

import "os"

type Config struct {
	ServiceName 		string 
	GRPCPort 			string 
	RedisURL    		string 
	KafkaBroker 		string
	KafkaInputTopic  	string 
	KafkaOutputTopic 	string 
	KafkaLogTopic 		string 
	AdminDBURL  		string 
}

func Load() *Config {
	return &Config{
		ServiceName: os.Getenv("SERVICE_NAME"),
		GRPCPort: os.Getenv("GRPC_PORT"),
		RedisURL: os.Getenv("REDIS_URL"),
		KafkaBroker: os.Getenv("KAFKA_BROKER"),
		KafkaInputTopic: os.Getenv("KAFKA_INPUT_TOPIC"),
		AdminDBURL: os.Getenv("ADMIN_DB_URL"),
		KafkaOutputTopic: os.Getenv("KAFKA_OUTPUT_TOPIC"),
		KafkaLogTopic: os.Getenv("KAFKA_LOG_TOPIC"),
	}
}
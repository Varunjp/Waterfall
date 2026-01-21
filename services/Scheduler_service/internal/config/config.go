package config

import "os"

type Config struct {
	KafkaBrokers  []string
	InputTopic    string
	OutputTopic   string
	ConsumerGroup string
}

func Load() *Config {
	return &Config{
		KafkaBrokers: []string{os.Getenv("KAFKA_BROKER")},
		InputTopic: os.Getenv("KAFKA_INPUT_TOPIC"),
		OutputTopic: os.Getenv("KAFKA_OUTPUT_TOPIC"),
		ConsumerGroup: os.Getenv("KAFKA_GROUP"),
	}
}
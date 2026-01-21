package config

import "os"

type Config struct {
	DBURL        string
	KafkaBroker  []string
	KafkaTopic   string
	PollInterval int
}

func Load() *Config {
	return &Config{
		DBURL: os.Getenv("DATABASE_URL"),
		KafkaBroker: []string{os.Getenv("KAFKA_BROKER")},
		KafkaTopic: os.Getenv("KAFKA_TOPIC"),
		PollInterval: 20,
	}
}
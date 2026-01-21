package config

import "os"

type Config struct {
	KafkaBroker []string
	KafkaTopic  string
	DBURL       string
	GroupID     string
}

func Load() *Config {
	return &Config{
		KafkaBroker: []string{os.Getenv("KAFKA_BROKER")},
		KafkaTopic: os.Getenv("KAFKA_JOB_TOPIC"),
		DBURL: os.Getenv("DATABASE_URL"),
		GroupID: "job-consumer-group",
	}
}
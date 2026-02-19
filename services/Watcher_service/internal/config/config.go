package config

import (
	"os"
	"strconv"
)

type Config struct {
	DBURL         string
	KafkaBroker   []string
	KafkaTopic    string
	JobTopic      string
	JobGroupID    string
	JobRunTopic   string
	JobRunGroupID string
	PollInterval  int
}

func Load() *Config {
	pollInterval := 20
	if value := os.Getenv("WATCHER_POLL_INTERVAL_SECONDS"); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 {
			pollInterval = parsed
		}
	}

	return &Config{
		DBURL:       os.Getenv("DATABASE_URL"),
		KafkaBroker: []string{os.Getenv("KAFKA_BROKER")},
		KafkaTopic:  os.Getenv("KAFKA_TOPIC"),
		JobTopic:    os.Getenv("KAFKA_JOB_TOPIC"),
		JobGroupID:  os.Getenv("KAFKA_CONSUMER_GROUP"),
		JobRunTopic: firstNonEmpty(
			os.Getenv("KAFKA_RUN_TOPIC"),
			os.Getenv("KAFAK_RUN_TOPIC"),
		),
		JobRunGroupID: firstNonEmpty(
			os.Getenv("KAFKA_RUN_GROUP_ID"),
			os.Getenv("KAFAK_RUN_GROUPID"),
		),
		PollInterval: pollInterval,
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DBURL         	string
	DBADMINURL	  	string
	KafkaBroker   	[]string
	KafkaTopic    	string
	JobTopic      	string
	JobGroupID    	string
	JobRunTopic   	string
	JobStatusTopic 	string 
	JobUsageTopic  	string
	TestTopic   	string 
	TestScheduler   string 
	JobUsageGroupID string 
	JobRunGroupID 	string
	PollInterval  	int
	RedisAddr 		string 
	RedisPassword 	string 
	RedisDB  		int
}

func Load() *Config {
	_ = godotenv.Load()
	
	pollInterval := 20
	if value := os.Getenv("WATCHER_POLL_INTERVAL_SECONDS"); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 {
			pollInterval = parsed
		}
	}
	redisDB, _ := strconv.Atoi(os.Getenv("REDIS_DB"))
	return &Config{
		DBURL:       os.Getenv("DATABASE_URL"),
		DBADMINURL: os.Getenv("DATABASE_ADMIN_URL"),
		KafkaBroker: []string{os.Getenv("KAFKA_BROKER")},
		KafkaTopic:  os.Getenv("KAFKA_TOPIC"),
		JobTopic:    os.Getenv("KAFKA_JOB_TOPIC"),
		JobGroupID:  os.Getenv("KAFKA_CONSUMER_GROUP"),
		JobUsageTopic: os.Getenv("KAFKA_JOB_USAGE_TOPIC"),
		JobUsageGroupID: os.Getenv("KAFKA_JOB_USAGE_TOPIC"),
		JobRunTopic: os.Getenv("KAFKA_RUN_TOPIC"),
		JobStatusTopic: os.Getenv("KAFKA_JOB_STATUS_TOPIC"),
		TestTopic: os.Getenv("TEST_TOPIC"),
		TestScheduler: os.Getenv("TEST_SCHEDULER_QUEUE"),
		JobRunGroupID: firstNonEmpty(
			os.Getenv("KAFKA_RUN_GROUP_ID"),
			os.Getenv("KAFAK_RUN_GROUPID"),
		),
		PollInterval: pollInterval,
		RedisAddr: os.Getenv("REDIS_ADDR"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		RedisDB: redisDB,
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

package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Env        string
	Scheduler  SchedulerConfig
	Kafka      KafkaConfig
	Redis      RedisConfig
	AdminDB    AdminDBConfig
	GRPC       GRPCConfig
	Metrics    MetricsConfig
}

type SchedulerConfig struct {
	ServiceName string
}

type KafkaConfig struct {
	Brokers          string
	JobCreateTopic   string
	JobUpdateTopic   string
	JobStatusTopic   string 
	ConsumerGroup    string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type AdminDBConfig struct {
	DSN string
}

type GRPCConfig struct {
	Port string
}

type MetricsConfig struct {
	Enabled bool
	Port    string
}

func Load() Config {
	cfg := Config{
		Env: mustEnv("APP_ENV"),

		Scheduler: SchedulerConfig{
			ServiceName: mustEnv("SERVICE_NAME"),
		},

		Kafka: KafkaConfig{
			Brokers:        mustEnv("KAFKA_BROKER"),
			JobCreateTopic: mustEnv("KAFKA_JOB_CREATE_TOPIC"),
			JobUpdateTopic: mustEnv("KAFKA_JOB_UPDATE_TOPIC"),
			JobStatusTopic: mustEnv("KAFKA_JOB_STATUS_TOPIC"),
			ConsumerGroup:  mustEnv("KAFKA_CONSUMER_GROUP"),
		},

		Redis: RedisConfig{
			Addr:     mustEnv("REDIS_ADDR"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       mustEnvInt("REDIS_DB"),
		},

		AdminDB: AdminDBConfig{
			DSN: mustEnv("ADMIN_DB_DSN"),
		},

		GRPC: GRPCConfig{
			Port: mustEnv("GRPC_PORT"),
		},

		Metrics: MetricsConfig{
			Enabled: mustEnvBool("METRICS_ENABLED"),
			Port:    mustEnv("METRICS_PORT"),
		},
	}

	return cfg
}

func mustEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic(fmt.Sprintf("missing required env var: %s", key))
	}
	return val
}

func getEnv(key, def string) string {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	return val
}

func mustEnvInt(key string) int {
	val := mustEnv(key)
	i, err := strconv.Atoi(val)
	if err != nil {
		panic(fmt.Sprintf("invalid int value for %s", key))
	}
	return i
}

func mustEnvBool(key string) bool {
	val := mustEnv(key)
	b, err := strconv.ParseBool(val)
	if err != nil {
		panic(fmt.Sprintf("invalid bool value for %s", key))
	}
	return b
}
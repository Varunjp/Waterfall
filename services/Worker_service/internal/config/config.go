package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppID             string
	WorkerID          string
	JobTypes          []string
	RedisAddr         string
	SchedulerGRPC     string
	MaxConcurrency    int
	HeartbeatInterval time.Duration
}

func Load() *Config {
	jobTypesRaw := must("JOB_TYPES")
	jobTypes := strings.Split(jobTypesRaw, ",")

	return &Config{
		AppID:             must("APP_ID"),
		WorkerID:          must("WORKER_ID"),
		JobTypes:          jobTypes,
		RedisAddr:         must("REDIS_ADDR"),
		SchedulerGRPC:     must("SCHEDULER_GRPC_ADDR"),
		MaxConcurrency:    intFromEnv("WORKER_MAX_CONCURRENCY", len(jobTypes)),
		HeartbeatInterval: time.Duration(intFromEnv("WORKER_HEARTBEAT_INTERVAL_SEC", 10)) * time.Second,
	}
}

func must(k string) string {
	v := os.Getenv(k)
	if v == "" {
		panic("missing env: " + k)
	}
	return v
}

func intFromEnv(k string, def int) int {
	v := os.Getenv(k)
	if v == "" {
		return def
	}

	out, err := strconv.Atoi(v)
	if err != nil {
		panic("invalid int env: " + k)
	}
	return out
}

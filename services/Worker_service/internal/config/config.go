package config

import (
	"os"
	"strings"
)

type Config struct {
	AppID        string
	WorkerID     string
	JobTypes     []string
	RedisAddr    string
	SchedulerGRPC string
}

func Load() *Config {

	return &Config{
		AppID:        must("APP_ID"),
		WorkerID:     must("WORKER_ID"),
		JobTypes:     strings.Split(must("JOB_TYPES"), ","),
		RedisAddr:    must("REDIS_ADDR"),
		SchedulerGRPC: must("SCHEDULER_GRPC_ADDR"),
	}
}

func must(k string) string {
	v := os.Getenv(k)
	if v == "" {
		panic("missing env: " + k)
	}
	return v
}
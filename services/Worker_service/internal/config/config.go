package config

import (
	"os"
)

type Config struct {
	WorkerID   string
	AppID      string
	Capabilities []string

	SchedulerAddr string
	HeartbeatSec  int
}

func Load() *Config {

	return &Config{
		WorkerID:        os.Getenv("WORKER_ID"),
		AppID:           os.Getenv("APP_ID"),
		Capabilities:    []string{"email"}, // can parse CSV if needed
		SchedulerAddr:   os.Getenv("SCHEDULER_ADDR"),
		HeartbeatSec:    10,
	}
}
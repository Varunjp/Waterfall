package config

import "os"

type Config struct {
	SchedulerAddr string
	WorkerID      string
	AppID         string
	Capabilitis   []string
	HeartbeatSec  int
	MaxConcurrentJobs int 
}

func Load() *Config {
	return &Config{
		SchedulerAddr: os.Getenv("SCHEDULER_ADDR"),
		WorkerID: os.Getenv("WORKER_ID"),
		AppID: os.Getenv("APP_ID"),
		Capabilitis: []string{"email"},
		HeartbeatSec: 10,
		MaxConcurrentJobs: 5,
	}
}
package config

import "os"

type Config struct {
	Port                string
	JobServiceURL       string
	AdminServiceURL     string
	SchedulerServiceURL string
	BillingServiceURL   string
	RedisURL            string
	JWTSecret           string
	RateLimit           int
}

func Load() *Config {
	return &Config{
		Port:                getEnv("PORT", "8081"),
		JobServiceURL:       getEnv("JOB_SERVICE_URL", "localhost:50051"),
		AdminServiceURL:     getEnv("ADMIN_SERVICE_URL", ""),
		SchedulerServiceURL: getEnv("SCHEDULER_SERVICE_URL", "localhost:50052"),
		BillingServiceURL:   getEnv("BILLING_SERVICE_URL", ""),
		RedisURL:            getEnv("REDIS_URL", "localhost:6379"),
		JWTSecret:           getEnv("JWT_SECRET", "secret"),
		RateLimit:           100,
	}
}

func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}

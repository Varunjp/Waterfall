package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	ServiceName 	string
	KafkaBrokers 	[]string
	KafkaTopic		string 
	DBDSN        	string
	DBADMINDNS		string 
	JWTKey			string 
	PORT 			string 
	Topic 		 	string 
	RedisAddr 		string 
	RedisPassword 	string 
	RedisDB  		int
}

func Load() *Config {
	redisDB, _ := strconv.Atoi(os.Getenv("REDIS_DB"))
	return &Config{
		ServiceName: "job-service",
		KafkaBrokers: []string{os.Getenv("KAFKA_BROKER")},
		KafkaTopic: os.Getenv("KAFKA_JOB_TOPIC"),
		JWTKey: os.Getenv("JWT_KEY"),
		PORT: os.Getenv("JOB_SERVICE_PORT"),
		DBDSN: must("DB_DSN"),
		DBADMINDNS: must("DB_ADMIN_DNS"),
		Topic: must("QUEUE_TOPIC"),
		RedisAddr: os.Getenv("REDIS_ADDR"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		RedisDB: redisDB,
	}
}

func must(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("missing env var %s", key)
	}
	return v
}
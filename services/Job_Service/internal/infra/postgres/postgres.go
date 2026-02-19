package postgresclient

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func MustConntect(dsn string) *pgxpool.Pool {
	cfg,err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatalf("invalid postgres dsn: %v",err)
	}

	cfg.MaxConns = 20 
	cfg.MinConns = 5
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.MaxConnIdleTime = 5 * time.Minute
	cfg.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(),cfg)
	if err != nil {
		log.Fatalf("postgres connection failed: %v",err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("postgres ping failed: %v",err)
	}

	log.Println("postgre connected")
	return pool 
}
package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx"
	"github.com/redis/go-redis/v9"
)

type AdminRepo struct {
	db *pgx.Conn
	rd *redis.Client
}

func NewAdminDb(dsn string,rd *redis.Client) *AdminRepo {
	cfg,err := pgx.ParseConnectionString(dsn)
	if err != nil {
		panic(err)
	}
	db,_ := pgx.Connect(cfg)
	return &AdminRepo{db: db,rd: rd}
}

func (r *AdminRepo) Concurrency(ctx context.Context,appID string)(int,error) {

	planKey := fmt.Sprintf("plan:%s",appID)
	limit,err := r.rd.Get(ctx,planKey).Int()
	
	if err != nil {
		return 0,err 
	}

	usageKey := fmt.Sprintf("usage:%s:%s",appID,time.Now().Format("2006-01"))
	usage,err := r.rd.Get(ctx,usageKey).Int()

	if err != nil {
		return 0,err
	}

	if usage >= limit {
		return 0,nil 
	}

	return limit,nil  
}

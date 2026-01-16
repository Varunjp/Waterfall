package idempotency

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Store struct {
	rdb *redis.Client 
}

func New(rdb *redis.Client) *Store {
	return &Store{rdb}
}

func (s *Store) IsProcessed(ctx context.Context, key string)(bool,error) {
	result, err := s.rdb.Exists(ctx,key).Result()
	return result > 0, err
}

func (s *Store) MarkProcessed(ctx context.Context, key string) error {
	return s.rdb.Set(ctx,key,"1",24*time.Hour).Err()
}
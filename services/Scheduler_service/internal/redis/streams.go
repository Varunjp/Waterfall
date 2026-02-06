package redisClient

import (
	"context"
	"strings"

	"github.com/redis/go-redis/v9"
)

func EnsureGroup(
	ctx context.Context,
	rdb *redis.Client,
	stream string,
	group string,
) error {

	err := rdb.XGroupCreateMkStream(
		ctx,
		stream,
		group,
		"$",
	).Err()

	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		return err
	}

	return nil 
}
package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func RateLimitMiddleware(rdb *redis.Client, limit int) gin.HandlerFunc {
	return func(c *gin.Context) {

		ctx := context.Background()
		key := "rate:" + c.ClientIP()

		count, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			c.Next()
			return
		}

		if count == 1 {
			rdb.Expire(ctx, key, time.Minute)
		}

		if int(count) > limit {
			c.AbortWithStatusJSON(429, gin.H{
				"error": "rate limit exceeded",
			})
			return
		}
		c.Next()
	}
}
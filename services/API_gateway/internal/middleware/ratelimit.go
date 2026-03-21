package middleware

import (
	"context"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func RateLimitMiddleware(rdb *redis.Client, limit int) gin.HandlerFunc {
	return func(c *gin.Context) {

		path := c.Request.URL.Path
		ctx := context.Background()

		if strings.HasPrefix(path,"/api/v1/jobs") && c.Request.Method == "POST" {
			key := "rate:jobs:" + c.ClientIP()
			count, _ := rdb.Incr(ctx, key).Result()
			if count == 1 {
				rdb.Expire(ctx, key, time.Second)
			}
			if count > 1000 { 
				c.AbortWithStatusJSON(429, gin.H{"error": "too many jobs"})
				return
			}
			c.Next()
			return
		}

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
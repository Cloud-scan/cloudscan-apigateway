package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

// RateLimiter middleware implements rate limiting using Redis
func RateLimiter(redisClient *redis.Client, requestsPerMinute int) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get user ID from context (if authenticated)
			userID := c.Get("user_id")
			var key string

			if userID != nil {
				// Rate limit by user ID
				key = fmt.Sprintf("rate_limit:user:%s", userID)
			} else {
				// Rate limit by IP address
				ip := c.RealIP()
				key = fmt.Sprintf("rate_limit:ip:%s", ip)
			}

			ctx := context.Background()

			// Increment counter
			count, err := redisClient.Incr(ctx, key).Result()
			if err != nil {
				log.WithError(err).Error("Failed to increment rate limit counter")
				// Allow request if Redis fails (fail open)
				return next(c)
			}

			// Set expiration on first request
			if count == 1 {
				redisClient.Expire(ctx, key, time.Minute)
			}

			// Check if limit exceeded
			if count > int64(requestsPerMinute) {
				return echo.NewHTTPError(http.StatusTooManyRequests, "rate limit exceeded")
			}

			// Set rate limit headers
			c.Response().Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", requestsPerMinute))
			c.Response().Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", requestsPerMinute-int(count)))

			return next(c)
		}
	}
}
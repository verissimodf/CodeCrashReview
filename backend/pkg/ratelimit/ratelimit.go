package ratelimit

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	client   *redis.Client
	requests int
	window   int // seconds
}

func NewRateLimiter(client *redis.Client, requests, window int) *RateLimiter {
	return &RateLimiter{
		client:   client,
		requests: requests,
		window:   window,
	}
}

func (r *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := r.getKey(c)
		allowed, retryAfter, err := r.allow(key)
		if err != nil {
			// Fail open on Redis errors
			c.Next()
			return
		}

		if !allowed {
			c.Header("X-RateLimit-Limit", strconv.Itoa(r.requests))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("Retry-After", strconv.Itoa(retryAfter))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate limit exceeded",
				"retry_after": retryAfter,
			})
			return
		}

		c.Header("X-RateLimit-Limit", strconv.Itoa(r.requests))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(r.getRemaining(key)))
		c.Next()
	}
}

func (r *RateLimiter) getKey(c *gin.Context) string {
	// Use IP + User ID if authenticated
	ip := c.ClientIP()
	if userID, exists := c.Get("user_id"); exists {
		return "ratelimit:" + userID.(string)
	}
	return "ratelimit:ip:" + ip
}

func (r *RateLimiter) allow(key string) (bool, int, error) {
	ctx := context.Background()
	now := time.Now().Unix()
	windowStart := now - int64(r.window)

	// Remove expired entries
	r.client.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart, 10))

	// Count current requests
	count, err := r.client.ZCard(ctx, key).Result()
	if err != nil {
		return false, 0, err
	}

	if count >= int64(r.requests) {
		// Get oldest entry to calculate retry-after
		oldest, err := r.client.ZRange(ctx, key, 0, 0).Result()
		if err != nil || len(oldest) == 0 {
			return false, r.window, nil
		}
		oldestTime, _ := strconv.ParseInt(oldest[0], 10, 64)
		retryAfter := int(oldestTime + int64(r.window) - now)
		if retryAfter < 1 {
			retryAfter = 1
		}
		return false, retryAfter, nil
	}

	// Add current request
	member := strconv.FormatInt(now, 10) + "-" + strconv.FormatInt(now%1000000, 10)
	r.client.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: member})
	r.client.Expire(ctx, key, time.Duration(r.window)*time.Second)

	return true, 0, nil
}

func (r *RateLimiter) getRemaining(key string) int {
	ctx := context.Background()
	now := time.Now().Unix()
	windowStart := now - int64(r.window)

	r.client.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart, 10))
	count, _ := r.client.ZCard(ctx, key).Result()

	remaining := r.requests - int(count)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// RateLimiterMiddleware creates a Gin middleware from config
func RateLimiterMiddleware(client *redis.Client, requests, window int) gin.HandlerFunc {
	limiter := NewRateLimiter(client, requests, window)
	return limiter.Middleware()
}

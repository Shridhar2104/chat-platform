package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Shridhar2104/chat-platform/auth-service/internal/models"
	"github.com/Shridhar2104/chat-platform/shared/database"
	"github.com/gin-gonic/gin"
)

type RedisTokenBucket struct{
	redis *database.RedisClient
	capacity int
	refillRate	float64
	tokensPerReq	int
	keyPrefix 	string
}

type BucketState struct{
	Tokens 		float64		`json:"tokens"`
	LastRefill	time.Time	`json:"last_refill"`
	Capacity 	int			`json:"capacity"`
	RefillRate	float64		`json:"refill_rate"`

}

func NewRedisTokenBucket(redis *database.RedisClient, capacity int, refillRate float64, tokensPerReq int) *RedisTokenBucket {
    return &RedisTokenBucket{
        redis:        redis,
        capacity:     capacity,
        refillRate:   refillRate,
        tokensPerReq: tokensPerReq,
        keyPrefix:    "rate_limit:bucket",
    }
}

func TokenBucketMiddleware(redis *database.RedisClient, requestsPerMinute int)	gin.HandlerFunc{
	//Convert requests per minute to tokens per second
	refillRate:= float64(requestsPerMinute)/60.0
	capacity := requestsPerMinute
	bucket := NewRedisTokenBucket(redis, capacity, refillRate, 1)
	return func(c *gin.Context){
		identifier := getIdentifier(c)
		allowed, tokens, waitTime, err:= bucket.AllowRequest(identifier)
		if err != nil {
            // If Redis is down, allow request but log error
            setRateLimitHeaders(c, capacity, capacity, time.Now())
            c.Next()
            return
        }
		setRateLimitHeaders(c, capacity, int(tokens), time.Now().Add(waitTime))

		if !allowed {
            c.JSON(http.StatusTooManyRequests, models.ErrorResponse{
                Error:   "rate_limit_exceeded",
                Message: fmt.Sprintf("Rate limit exceeded. Try again in %.2f seconds.", waitTime.Seconds()),
            })
            c.Abort()
            return
        }

        c.Next()
	}

}
// AllowRequest checks if a request is allowed and consumes tokens
func (tb *RedisTokenBucket) AllowRequest(identifier string) (allowed bool, remainingTokens float64, waitTime time.Duration, err error) {
    return tb.AllowRequestWithTokens(identifier, tb.tokensPerReq)
}


func (tb *RedisTokenBucket) AllowRequestWithTokens(identifier string, tokensNeeded int)(allowed bool,remainingTokens float64, waitTime time.Duration, err error){
	ctx := context.Background()
    key := fmt.Sprintf("%s:%s", tb.keyPrefix, identifier)
    now := time.Now()

	//lua script for atomic token bucket ops

	luaScript:= `
		local key = KEYS[1]
		local capacity = tonumber(ARGV[1])
		local refill_rate = tonumber(ARGV[2])
		local tokens_needed = tonumber(ARGV[3])
		local now = tonumber(ARGV[4])

		-- Get current bucket state
		local bucket_data = redis.call('HMGET', key, 'tokens', 'last_refill')
		local current_tokens = tonumber(bucket_data[1]) or capacity
		local last_refill = tonumber(bucket_data[2]) or now

		-- Calculate tokens to add based on time elapsed
		local time_elapsed = now - last_refill
		local tokens_to_add = time_elapsed * refill_rate

		-- Update tokens (cap at capacity)
		current_tokens = math.min(capacity, current_tokens + tokens_to_add)

		-- Check if enough tokens available
		if current_tokens >= tokens_needed then
			-- Consume tokens
			current_tokens = current_tokens - tokens_needed
			
			-- Update bucket state
			redis.call('HMSET', key, 'tokens', current_tokens, 'last_refill', now)
			redis.call('EXPIRE', key, 300) -- 5 minute expiry for cleanup
			
			return {1, current_tokens, 0} -- allowed, remaining_tokens, wait_time
		else
			-- Not enough tokens - calculate wait time
			local tokens_deficit = tokens_needed - current_tokens
			local wait_time = tokens_deficit / refill_rate
			
			-- Update last_refill time but don't consume tokens
			redis.call('HMSET', key, 'tokens', current_tokens, 'last_refill', now)
			redis.call('EXPIRE', key, 300)
			
			return {0, current_tokens, wait_time} -- not allowed, remaining_tokens, wait_time
		end
	`

	result, err := tb.redis.Client.Eval(ctx, luaScript, []string{key}, 
        tb.capacity, tb.refillRate, tokensNeeded, now.Unix()).Result()
    if err != nil {
        return false, 0, 0, fmt.Errorf("redis eval error: %w", err)
    }

    // Parse result
    resultSlice, ok := result.([]interface{})
    if !ok || len(resultSlice) != 3 {
        return false, 0, 0, fmt.Errorf("unexpected redis result format")
    }

    allowedInt := resultSlice[0].(int64)
    remainingTokens = resultSlice[1].(float64)
    waitTimeSeconds := resultSlice[2].(float64)
    waitTime = time.Duration(waitTimeSeconds * float64(time.Second))

    return allowedInt == 1, remainingTokens, waitTime, nil

}

// GetBucketState returns current state of the bucket
func (tb *RedisTokenBucket) GetBucketState(identifier string) (*BucketState, error) {
    ctx := context.Background()
    key := fmt.Sprintf("%s:%s", tb.keyPrefix, identifier)
    
    data, err := tb.redis.Client.HMGet(ctx, key, "tokens", "last_refill").Result()
    if err != nil {
        return nil, err
    }
    
    now := time.Now()
    
    var tokens float64 = float64(tb.capacity) // default
    var lastRefill time.Time = now           // default
    
    if data[0] != nil {
        if tokensStr, ok := data[0].(string); ok {
            tokens, _ = strconv.ParseFloat(tokensStr, 64)
        }
    }
    
    if data[1] != nil {
        if refillStr, ok := data[1].(string); ok {
            if timestamp, err := strconv.ParseInt(refillStr, 10, 64); err == nil {
                lastRefill = time.Unix(timestamp, 0)
            }
        }
    }
    
    return &BucketState{
        Tokens:     tokens,
        LastRefill: lastRefill,
        Capacity:   tb.capacity,
        RefillRate: tb.refillRate,
    }, nil
}

// Reset clears the bucket for an identifier
func (tb *RedisTokenBucket) Reset(identifier string) error {
    ctx := context.Background()
    key := fmt.Sprintf("%s:%s", tb.keyPrefix, identifier)
    return tb.redis.Client.Del(ctx, key).Err()
}
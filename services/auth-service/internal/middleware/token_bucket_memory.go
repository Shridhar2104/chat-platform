package middleware

import (
    "fmt"
    "math"
    "net/http"
    "sync"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/Shridhar2104/chat-platform/auth-service/internal/models"
)

type MemoryTokenBucket struct {
    buckets      map[string]*tokenBucket
    mutex        sync.RWMutex
    capacity     int
    refillRate   float64
    tokensPerReq int
}

type tokenBucket struct {
    tokens     float64
    lastRefill time.Time
    mutex      sync.Mutex
}

func NewMemoryTokenBucket(capacity int, refillRate float64, tokensPerReq int) *MemoryTokenBucket {
    tb := &MemoryTokenBucket{
        buckets:      make(map[string]*tokenBucket),
        capacity:     capacity,
        refillRate:   refillRate,
        tokensPerReq: tokensPerReq,
    }
    
    // Start cleanup routine
    go tb.cleanup()
    
    return tb
}

// MemoryTokenBucketMiddleware creates an in-memory token bucket rate limiter
func MemoryTokenBucketMiddleware(requestsPerMinute int) gin.HandlerFunc {
    refillRate := float64(requestsPerMinute) / 60.0
    capacity := requestsPerMinute
    
    limiter := NewMemoryTokenBucket(capacity, refillRate, 1)

    return func(c *gin.Context) {
        identifier := getIdentifier(c)
        
        allowed, tokens, waitTime := limiter.AllowRequest(identifier)
        
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

// AllowRequest checks if request is allowed
func (mtb *MemoryTokenBucket) AllowRequest(identifier string) (allowed bool, remainingTokens float64, waitTime time.Duration) {
    return mtb.AllowRequestWithTokens(identifier, mtb.tokensPerReq)
}

// AllowRequestWithTokens allows custom token cost
func (mtb *MemoryTokenBucket) AllowRequestWithTokens(identifier string, tokensNeeded int) (allowed bool, remainingTokens float64, waitTime time.Duration) {
    bucket := mtb.getOrCreateBucket(identifier)
    
    bucket.mutex.Lock()
    defer bucket.mutex.Unlock()
    
    now := time.Now()
    
    // Calculate tokens to add
    timeElapsed := now.Sub(bucket.lastRefill).Seconds()
    tokensToAdd := timeElapsed * mtb.refillRate
    
    // Update tokens (cap at capacity)
    bucket.tokens = math.Min(float64(mtb.capacity), bucket.tokens+tokensToAdd)
    bucket.lastRefill = now
    
    // Check if enough tokens
    if bucket.tokens >= float64(tokensNeeded) {
        bucket.tokens -= float64(tokensNeeded)
        return true, bucket.tokens, 0
    }
    
    // Calculate wait time
    deficit := float64(tokensNeeded) - bucket.tokens
    waitTime = time.Duration(deficit/mtb.refillRate) * time.Second
    
    return false, bucket.tokens, waitTime
}

// getOrCreateBucket safely gets or creates a bucket
func (mtb *MemoryTokenBucket) getOrCreateBucket(identifier string) *tokenBucket {
    mtb.mutex.RLock()
    bucket, exists := mtb.buckets[identifier]
    mtb.mutex.RUnlock()
    
    if exists {
        return bucket
    }
    
    mtb.mutex.Lock()
    defer mtb.mutex.Unlock()
    
    // Double check after acquiring write lock
    if bucket, exists := mtb.buckets[identifier]; exists {
        return bucket
    }
    
    // Create new bucket
    bucket = &tokenBucket{
        tokens:     float64(mtb.capacity),
        lastRefill: time.Now(),
    }
    mtb.buckets[identifier] = bucket
    
    return bucket
}

// GetBucketState returns current bucket state
func (mtb *MemoryTokenBucket) GetBucketState(identifier string) (tokens float64, lastRefill time.Time) {
    bucket := mtb.getOrCreateBucket(identifier)
    
    bucket.mutex.Lock()
    defer bucket.mutex.Unlock()
    
    // Update tokens before returning state
    now := time.Now()
    timeElapsed := now.Sub(bucket.lastRefill).Seconds()
    tokensToAdd := timeElapsed * mtb.refillRate
    bucket.tokens = math.Min(float64(mtb.capacity), bucket.tokens+tokensToAdd)
    bucket.lastRefill = now
    
    return bucket.tokens, bucket.lastRefill
}

// cleanup removes old buckets
func (mtb *MemoryTokenBucket) cleanup() {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        mtb.mutex.Lock()
        now := time.Now()
        
        for identifier, bucket := range mtb.buckets {
            bucket.mutex.Lock()
            // Remove buckets that haven't been used in 10 minutes
            if now.Sub(bucket.lastRefill) > 10*time.Minute {
                delete(mtb.buckets, identifier)
            }
            bucket.mutex.Unlock()
        }
        
        mtb.mutex.Unlock()
    }
}
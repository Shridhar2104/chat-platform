package middleware

import (
    "fmt"
    "strconv"
    "time"

    "github.com/gin-gonic/gin"
)

// getIdentifier determines the rate limiting identifier
func getIdentifier(c *gin.Context) string {
    // Priority: User ID > IP Address
    if userID := c.GetString("user_id"); userID != "" {
        return fmt.Sprintf("user:%s", userID)
    }
    return fmt.Sprintf("ip:%s", c.ClientIP())
}

// setRateLimitHeaders sets standard rate limiting headers
func setRateLimitHeaders(c *gin.Context, limit int, remaining int, resetTime time.Time) {
    c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
    c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
    c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))
    c.Header("X-RateLimit-Type", "token-bucket")
}

// getEndpointIdentifier creates endpoint-specific identifiers
func getEndpointIdentifier(c *gin.Context, baseIdentifier string) string {
    endpoint := c.FullPath()
    return fmt.Sprintf("%s:%s", baseIdentifier, endpoint)
}
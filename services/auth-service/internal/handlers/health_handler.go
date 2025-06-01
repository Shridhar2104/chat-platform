package handlers

import (
    "context"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/Shridhar2104/chat-platform/shared/database"
)

type HealthHandler struct {
    db    *database.PostgresDB
    redis *database.RedisClient
}

type HealthResponse struct {
    Status    string                 `json:"status"`
    Timestamp string                 `json:"timestamp"`
    Service   string                 `json:"service"`
    Version   string                 `json:"version,omitempty"`
    Checks    map[string]HealthCheck `json:"checks"`
}

type HealthCheck struct {
    Status      string `json:"status"`
    Message     string `json:"message,omitempty"`
    ResponseTime string `json:"response_time,omitempty"`
}

const (
    StatusHealthy   = "healthy"
    StatusUnhealthy = "unhealthy"
    StatusDegraded  = "degraded"
)

func NewHealthHandler(db *database.PostgresDB, redis *database.RedisClient) *HealthHandler {
    return &HealthHandler{
        db:    db,
        redis: redis,
    }
}

// HealthCheck provides a simple health check endpoint
func (h *HealthHandler) HealthCheck(c *gin.Context) {
    start := time.Now()
    
    // Check database connectivity
    dbStatus := h.checkDatabase()
    
    // Check Redis connectivity
    redisStatus := h.checkRedis()
    
    // Determine overall status
    overallStatus := h.determineOverallStatus(dbStatus, redisStatus)
    
    response := HealthResponse{
        Status:    overallStatus,
        Timestamp: time.Now().UTC().Format(time.RFC3339),
        Service:   "auth-service",
        Version:   "v1.0.0", // You can make this configurable
        Checks: map[string]HealthCheck{
            "database": dbStatus,
            "redis":    redisStatus,
        },
    }
    
    // Set HTTP status based on health
    statusCode := http.StatusOK
    if overallStatus == StatusUnhealthy {
        statusCode = http.StatusServiceUnavailable
    } else if overallStatus == StatusDegraded {
        statusCode = http.StatusOK // 200 but with degraded status
    }
    
    // Add total response time
    response.Checks["overall"] = HealthCheck{
        Status:       overallStatus,
        ResponseTime: time.Since(start).String(),
    }
    
    c.JSON(statusCode, response)
}

// LivenessProbe for Kubernetes liveness probe
func (h *HealthHandler) LivenessProbe(c *gin.Context) {
    // Simple check - if the service is running, it's alive
    c.JSON(http.StatusOK, gin.H{
        "status": "alive",
        "timestamp": time.Now().UTC().Format(time.RFC3339),
    })
}

// ReadinessProbe for Kubernetes readiness probe
func (h *HealthHandler) ReadinessProbe(c *gin.Context) {
    // Check if service is ready to handle requests
    dbStatus := h.checkDatabase()
    
    if dbStatus.Status == StatusHealthy {
        c.JSON(http.StatusOK, gin.H{
            "status": "ready",
            "timestamp": time.Now().UTC().Format(time.RFC3339),
        })
    } else {
        c.JSON(http.StatusServiceUnavailable, gin.H{
            "status": "not_ready",
            "timestamp": time.Now().UTC().Format(time.RFC3339),
            "reason": "database_unavailable",
        })
    }
}

// checkDatabase verifies database connectivity
func (h *HealthHandler) checkDatabase() HealthCheck {
    start := time.Now()
    
    // Create a context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    // Ping the database
    err := h.db.Ping()
    responseTime := time.Since(start)
    
    if err != nil {
        return HealthCheck{
            Status:       StatusUnhealthy,
            Message:      "Database connection failed: " + err.Error(),
            ResponseTime: responseTime.String(),
        }
    }
    
    // Additional check: Try a simple query
    var result int
    err = h.db.DB.GetContext(ctx, &result, "SELECT 1")
    totalResponseTime := time.Since(start)
    
    if err != nil {
        return HealthCheck{
            Status:       StatusUnhealthy,
            Message:      "Database query failed: " + err.Error(),
            ResponseTime: totalResponseTime.String(),
        }
    }
    
    // Check response time - if too slow, mark as degraded
    if totalResponseTime > 2*time.Second {
        return HealthCheck{
            Status:       StatusDegraded,
            Message:      "Database responding slowly",
            ResponseTime: totalResponseTime.String(),
        }
    }
    
    return HealthCheck{
        Status:       StatusHealthy,
        Message:      "Database connection successful",
        ResponseTime: totalResponseTime.String(),
    }
}

// checkRedis verifies Redis connectivity
func (h *HealthHandler) checkRedis() HealthCheck {
    start := time.Now()
    
    // Create a context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    // Ping Redis
    pong, err := h.redis.Client.Ping(ctx).Result()
    responseTime := time.Since(start)
    
    if err != nil {
        return HealthCheck{
            Status:       StatusUnhealthy,
            Message:      "Redis connection failed: " + err.Error(),
            ResponseTime: responseTime.String(),
        }
    }
    
    if pong != "PONG" {
        return HealthCheck{
            Status:       StatusUnhealthy,
            Message:      "Redis ping returned unexpected response: " + pong,
            ResponseTime: responseTime.String(),
        }
    }
    
    // Additional check: Try a simple operation
    testKey := "health_check_" + time.Now().Format("20060102150405")
    err = h.redis.Client.Set(ctx, testKey, "test", 10*time.Second).Err()
    if err != nil {
        return HealthCheck{
            Status:       StatusDegraded,
            Message:      "Redis set operation failed: " + err.Error(),
            ResponseTime: time.Since(start).String(),
        }
    }
    
    // Clean up test key
    h.redis.Client.Del(ctx, testKey)
    totalResponseTime := time.Since(start)
    
    // Check response time
    if totalResponseTime > 1*time.Second {
        return HealthCheck{
            Status:       StatusDegraded,
            Message:      "Redis responding slowly",
            ResponseTime: totalResponseTime.String(),
        }
    }
    
    return HealthCheck{
        Status:       StatusHealthy,
        Message:      "Redis connection successful",
        ResponseTime: totalResponseTime.String(),
    }
}

// determineOverallStatus calculates the overall service health
func (h *HealthHandler) determineOverallStatus(dbStatus, redisStatus HealthCheck) string {
    // Database is critical - if it's down, service is unhealthy
    if dbStatus.Status == StatusUnhealthy {
        return StatusUnhealthy
    }
    
    // If database is degraded or Redis has issues, service is degraded
    if dbStatus.Status == StatusDegraded || redisStatus.Status != StatusHealthy {
        return StatusDegraded
    }
    
    // All good
    return StatusHealthy
}

// MetricsEndpoint provides basic metrics (optional)
func (h *HealthHandler) MetricsEndpoint(c *gin.Context) {
    // You can add custom metrics here
    // For now, just basic info
    
    metrics := gin.H{
        "service": "auth-service",
        "uptime":  time.Since(startTime).String(),
        "timestamp": time.Now().UTC().Format(time.RFC3339),
        "go_version": "1.21", // You can get this dynamically
        "build_info": gin.H{
            "version": "v1.0.0",
            "commit":  "unknown", // Set this during build
            "date":    "unknown", // Set this during build
        },
    }
    
    c.JSON(http.StatusOK, metrics)
}

// Package-level variable to track start time
var startTime = time.Now()
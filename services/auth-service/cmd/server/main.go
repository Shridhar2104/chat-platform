package main

import (
    "context"
 
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/Shridhar2104/chat-platform/shared/config"
    "github.com/Shridhar2104/chat-platform/shared/database"
    "github.com/Shridhar2104/chat-platform/auth-service/internal/handlers"
    "github.com/Shridhar2104/chat-platform/auth-service/internal/middleware"
    "github.com/Shridhar2104/chat-platform/auth-service/internal/repository"
    "github.com/Shridhar2104/chat-platform/auth-service/internal/services"
)

func main() {
    // Load configuration
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // Setup database connections
    db, err := database.NewPostgresConnection(cfg.DatabaseURL)
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    defer db.Close()

    redis, err := database.NewRedisConnection(cfg.RedisURL)
    if err != nil {
        log.Fatalf("Failed to connect to Redis: %v", err)
    }
    defer redis.Close()

    // Initialize repositories
    userRepo := repository.NewUserRepository(db)

    // Initialize services
    jwtService := services.NewJWTService(cfg.JWTSecret, cfg.JWTExpiration, cfg.RefreshExpiration)
    authService := services.NewAuthService(userRepo, jwtService, redis)

    // Initialize handlers
    authHandler := handlers.NewAuthHandler(authService)
    healthHandler := handlers.NewHealthHandler(db, redis)

    // Setup router
    router := setupRouter(cfg, authHandler, healthHandler, redis)

    // Setup server
    srv := &http.Server{
        Addr:    ":" + cfg.Port,
        Handler: router,
    }

    // Start server in a goroutine
    go func() {
        log.Printf("Auth service starting on port %s", cfg.Port)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Failed to start server: %v", err)
        }
    }()

    // Wait for interrupt signal to gracefully shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    log.Println("Shutting down server...")

    // Graceful shutdown with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        log.Fatalf("Server forced to shutdown: %v", err)
    }

    log.Println("Server exited")
}

func setupRouter(cfg *config.Config, authHandler *handlers.AuthHandler, healthHandler *handlers.HealthHandler, redis *database.RedisClient) *gin.Engine {
    if cfg.Environment == "production" {
        gin.SetMode(gin.ReleaseMode)
    }

    router := gin.New()
    
    // Global middleware
    router.Use(gin.Logger())
    router.Use(gin.Recovery())
    

    // Health check endpoints (no rate limiting)
    router.GET("/health", healthHandler.HealthCheck)
    router.GET("/health/live", healthHandler.LivenessProbe)
    router.GET("/health/ready", healthHandler.ReadinessProbe)
    router.GET("/metrics", healthHandler.MetricsEndpoint)

    // API v1 routes
    v1 := router.Group("/api/v1")
    
    // Simple Token Bucket Rate Limiting (1 token per request for all endpoints)
    if cfg.RateLimitEnabled {
        if cfg.Environment == "production" || cfg.Environment == "staging" {
            // Redis-based token bucket for production
            v1.Use(middleware.TokenBucketMiddleware(redis, cfg.RateLimitRPM))
        } else {
            // Memory-based token bucket for development
            v1.Use(middleware.MemoryTokenBucketMiddleware(cfg.RateLimitRPM))
        }
    }

    // Auth routes
    auth := v1.Group("/auth")
    {
        auth.POST("/register", authHandler.Register)
        auth.POST("/login", authHandler.Login)
        auth.POST("/refresh", authHandler.RefreshToken)
        auth.POST("/forgot-password", authHandler.ForgotPassword)
        auth.POST("/reset-password", authHandler.ResetPassword)
    }

    // Protected routes
    protected := v1.Group("/auth")
    protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
    {
        protected.POST("/logout", authHandler.Logout)
        protected.GET("/me", authHandler.GetCurrentUser)
        protected.PUT("/change-password", authHandler.ChangePassword)
    }

    return router
}
package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
    // Server
    Port     string
    Environment string
    
    // Database
    DatabaseURL string
    RedisURL    string
    
    // Kafka
    KafkaBrokers []string
    
    // JWT
    JWTSecret         string
    JWTExpiration     time.Duration
    RefreshExpiration time.Duration
    
    // Azure
    AzureKeyVaultURL string
    
    // Region & GDPR
    Region     string
    GDPRRegion string
    
    // Rate Limiting
    RateLimitEnabled bool
    RateLimitRPM     int
}

func Load() (*Config, error) {
    // Load .env file if it exists
    envFiles := []string{".env", "../.env", "../../.env"}
    
    loaded := false
    for _, envFile := range envFiles {
        if err := godotenv.Load(envFile); err == nil {
            log.Printf("DEBUG: Successfully loaded .env from: %s", envFile)
            loaded = true
            break
        }
    }
    
    if !loaded {
        log.Printf("DEBUG: No .env file found, using system environment variables")
    }
    
    config := &Config{
        Port:        getEnv("PORT", "8080"),
        Environment: getEnv("ENVIRONMENT", "development"),
        
        DatabaseURL: getEnv("DATABASE_URL", "postgres://chatuser:password@localhost:5432/chatdb?sslmode=disable"),
        RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379"),
        
        JWTSecret:         getEnv("JWT_SECRET", "your-super-secret-jwt-key"),
        JWTExpiration:     getDurationEnv("JWT_EXPIRATION", 15*time.Minute),
        RefreshExpiration: getDurationEnv("REFRESH_EXPIRATION", 7*24*time.Hour),
        
        AzureKeyVaultURL: getEnv("AZURE_KEY_VAULT_URL", ""),
        
        Region:     getEnv("REGION", "us-east-1"),
        GDPRRegion: getEnv("GDPR_REGION", "us"),
        
        RateLimitEnabled: getBoolEnv("RATE_LIMIT_ENABLED", true),
        RateLimitRPM:     getIntEnv("RATE_LIMIT_RPM", 60),
    }
    
    // Parse Kafka brokers
    kafkaBrokers := getEnv("KAFKA_BROKERS", "localhost:9092")
    config.KafkaBrokers = []string{kafkaBrokers}
    
    return config, nil
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if intValue, err := strconv.Atoi(value); err == nil {
            return intValue
        }
    }
    return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
    if value := os.Getenv(key); value != "" {
        if boolValue, err := strconv.ParseBool(value); err == nil {
            return boolValue
        }
    }
    return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
    if value := os.Getenv(key); value != "" {
        if duration, err := time.ParseDuration(value); err == nil {
            return duration
        }
    }
    return defaultValue
}
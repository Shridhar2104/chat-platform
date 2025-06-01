package middleware

import (
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/Shridhar2104/chat-platform/auth-service/internal/models"
    "github.com/Shridhar2104/chat-platform/auth-service/internal/services"
)

func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
    jwtService := services.NewJWTService(jwtSecret, 0, 0) // Only need validation

    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, models.ErrorResponse{
                Error:   "missing_authorization",
                Message: "Authorization header is required",
            })
            c.Abort()
            return
        }

        // Check for Bearer token
        tokenParts := strings.Split(authHeader, " ")
        if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
            c.JSON(http.StatusUnauthorized, models.ErrorResponse{
                Error:   "invalid_authorization",
                Message: "Authorization header must be in format: Bearer <token>",
            })
            c.Abort()
            return
        }

        token := tokenParts[1]
        claims, err := jwtService.ValidateAccessToken(token)
        if err != nil {
            c.JSON(http.StatusUnauthorized, models.ErrorResponse{
                Error:   "invalid_token",
                Message: "Invalid or expired token",
            })
            c.Abort()
            return
        }

        // Set user context
        c.Set("user_id", claims.UserID.String())
        c.Set("email", claims.Email)
        c.Set("device_id", claims.DeviceID)

        c.Next()
    }
}
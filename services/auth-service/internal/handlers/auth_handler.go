package handlers

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/Shridhar2104/chat-platform/auth-service/internal/models"
    "github.com/Shridhar2104/chat-platform/auth-service/internal/services"
)

type AuthHandler struct {
    authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
    return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
    var req models.RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{
            Error:   "validation_error",
            Message: err.Error(),
        })
        return
    }

    user, accessToken, refreshToken, expiresAt, err := h.authService.Register(req.Email, req.Password, req.DisplayName)
    if err != nil {
        status := http.StatusInternalServerError
        if err.Error() == "email already registered" {
            status = http.StatusConflict
        }
        c.JSON(status, models.ErrorResponse{
            Error:   "registration_failed",
            Message: err.Error(),
        })
        return
    }

    response := models.AuthResponse{
        User: models.UserResponse{
            ID:            user.ID,
            Email:         user.Email,
            DisplayName:   user.DisplayName,
            AvatarURL:     user.AvatarURL,
            EmailVerified: user.EmailVerified,
            CreatedAt:     user.CreatedAt.Format(time.RFC3339),
        },
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        ExpiresAt:    expiresAt.Unix(),
    }

    c.JSON(http.StatusCreated, response)
}

func (h *AuthHandler) Login(c *gin.Context) {
    var req models.LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{
            Error:   "validation_error",
            Message: err.Error(),
        })
        return
    }

    user, accessToken, refreshToken, expiresAt, err := h.authService.Login(req.Email, req.Password, req.DeviceID)
    if err != nil {
        c.JSON(http.StatusUnauthorized, models.ErrorResponse{
            Error:   "login_failed",
            Message: "Invalid credentials",
        })
        return
    }

    response := models.AuthResponse{
        User: models.UserResponse{
            ID:            user.ID,
            Email:         user.Email,
            DisplayName:   user.DisplayName,
            AvatarURL:     user.AvatarURL,
            EmailVerified: user.EmailVerified,
            CreatedAt:     user.CreatedAt.Format(time.RFC3339),
        },
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        ExpiresAt:    expiresAt.Unix(),
    }

    c.JSON(http.StatusOK, response)
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
    var req models.RefreshTokenRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{
            Error:   "validation_error",
            Message: err.Error(),
        })
        return
    }

    accessToken, refreshToken, expiresAt, err := h.authService.RefreshToken(req.RefreshToken, req.DeviceID)
    if err != nil {
        c.JSON(http.StatusUnauthorized, models.ErrorResponse{
            Error:   "refresh_failed",
            Message: "Invalid or expired refresh token",
        })
        return
    }

    response := gin.H{
        "access_token":  accessToken,
        "refresh_token": refreshToken,
        "expires_at":    expiresAt.Unix(),
    }

    c.JSON(http.StatusOK, response)
}

func (h *AuthHandler) Logout(c *gin.Context) {
    userID := c.GetString("user_id")
    deviceID := c.Query("device_id")

    if deviceID == "" {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{
            Error:   "validation_error",
            Message: "device_id is required",
        })
        return
    }

    userUUID, err := uuid.Parse(userID)
    if err != nil {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{
            Error:   "invalid_user_id",
            Message: "Invalid user ID format",
        })
        return
    }

    err = h.authService.Logout(userUUID, deviceID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, models.ErrorResponse{
            Error:   "logout_failed",
            Message: err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, models.SuccessResponse{
        Message: "Logged out successfully",
    })
}

func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
    userID := c.GetString("user_id")
    userUUID, err := uuid.Parse(userID)
    if err != nil {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{
            Error:   "invalid_user_id",
            Message: "Invalid user ID format",
        })
        return
    }

    user, err := h.authService.GetUserByID(userUUID)
    if err != nil {
        c.JSON(http.StatusNotFound, models.ErrorResponse{
            Error:   "user_not_found",
            Message: "User not found",
        })
        return
    }

    response := models.UserResponse{
        ID:            user.ID,
        Email:         user.Email,
        DisplayName:   user.DisplayName,
        AvatarURL:     user.AvatarURL,
        EmailVerified: user.EmailVerified,
        CreatedAt:     user.CreatedAt.Format(time.RFC3339),
    }

    c.JSON(http.StatusOK, response)
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
    var req models.ChangePasswordRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{
            Error:   "validation_error",
            Message: err.Error(),
        })
        return
    }

    userID := c.GetString("user_id")
    userUUID, err := uuid.Parse(userID)
    if err != nil {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{
            Error:   "invalid_user_id",
            Message: "Invalid user ID format",
        })
        return
    }

    err = h.authService.ChangePassword(userUUID, req.CurrentPassword, req.NewPassword)
    if err != nil {
        status := http.StatusInternalServerError
        if err.Error() == "current password is incorrect" {
            status = http.StatusUnauthorized
        }
        c.JSON(status, models.ErrorResponse{
            Error:   "password_change_failed",
            Message: err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, models.SuccessResponse{
        Message: "Password changed successfully",
    })
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
    var req models.ForgotPasswordRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{
            Error:   "validation_error",
            Message: err.Error(),
        })
        return
    }

    // TODO: Implement password reset functionality
    // For now, return success (security through obscurity)
    c.JSON(http.StatusOK, models.SuccessResponse{
        Message: "If an account with that email exists, a password reset link has been sent",
    })
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
    var req models.ResetPasswordRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{
            Error:   "validation_error",
            Message: err.Error(),
        })
        return
    }

    // TODO: Implement password reset functionality
    c.JSON(http.StatusOK, models.SuccessResponse{
        Message: "Password reset successfully",
    })
}
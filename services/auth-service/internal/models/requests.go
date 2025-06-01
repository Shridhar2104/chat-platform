package models

import "github.com/google/uuid"


type RegisterRequest struct {
    Email       string `json:"email" binding:"required,email"`
    Password    string `json:"password" binding:"required,min=8"`
    DisplayName string `json:"display_name" binding:"required,min=2,max=100"`
}

type LoginRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required"`
    DeviceID string `json:"device_id" binding:"required"`
}

type RefreshTokenRequest struct {
    RefreshToken string `json:"refresh_token" binding:"required"`
    DeviceID     string `json:"device_id" binding:"required"`
}


type ForgotPasswordRequest struct {
    Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
    Token       string `json:"token" binding:"required"`
    NewPassword string `json:"new_password" binding:"required,min=8"`
}

type ChangePasswordRequest struct {
    CurrentPassword string `json:"current_password" binding:"required"`
    NewPassword     string `json:"new_password" binding:"required,min=8"`
}


type AuthResponse struct {
    User         UserResponse `json:"user"`
    AccessToken  string       `json:"access_token"`
    RefreshToken string       `json:"refresh_token"`
    ExpiresAt    int64        `json:"expires_at"`
}

type UserResponse struct {
    ID            uuid.UUID `json:"id"`
    Email         string    `json:"email"`
    DisplayName   string    `json:"display_name"`
    AvatarURL     *string   `json:"avatar_url"`
    EmailVerified bool      `json:"email_verified"`
    CreatedAt     string    `json:"created_at"`
}

type ErrorResponse struct {
    Error   string `json:"error"`
    Message string `json:"message,omitempty"`
}

type SuccessResponse struct {
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}
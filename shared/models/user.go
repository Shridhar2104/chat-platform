package models

import (
    "time"
    "github.com/google/uuid"
)

type User struct {
    ID            uuid.UUID  `json:"id" db:"id"`
    Email         string     `json:"email" db:"email"`
    PasswordHash  string     `json:"-" db:"password_hash"`
    DisplayName   string     `json:"display_name" db:"display_name"`
    AvatarURL     *string    `json:"avatar_url" db:"avatar_url"`
    EmailVerified bool       `json:"email_verified" db:"email_verified"`
    CreatedAt     time.Time  `json:"created_at" db:"created_at"`
    UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

type UserSession struct {
    ID               uuid.UUID `json:"id" db:"id"`
    UserID           uuid.UUID `json:"user_id" db:"user_id"`
    DeviceID         string    `json:"device_id" db:"device_id"`
    RefreshTokenHash string    `json:"-" db:"refresh_token_hash"`
    ExpiresAt        time.Time `json:"expires_at" db:"expires_at"`
    CreatedAt        time.Time `json:"created_at" db:"created_at"`
}

type UserProfile struct {
    UserID      uuid.UUID              `json:"user_id" db:"user_id"`
    Bio         *string                `json:"bio" db:"bio"`
    Timezone    *string                `json:"timezone" db:"timezone"`
    Status      string                 `json:"status" db:"status"`
    LastSeen    *time.Time             `json:"last_seen" db:"last_seen"`
    Preferences map[string]interface{} `json:"preferences" db:"preferences"`
}

type UserPresence struct {
    UserID        uuid.UUID `json:"user_id" db:"user_id"`
    DeviceID      string    `json:"device_id" db:"device_id"`
    Status        string    `json:"status" db:"status"`
    LastHeartbeat time.Time `json:"last_heartbeat" db:"last_heartbeat"`
}
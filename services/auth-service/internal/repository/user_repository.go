package repository

import (
    "database/sql"
    "fmt"
    "time"

    "github.com/google/uuid"
    "github.com/Shridhar2104/chat-platform/shared/database"
    "github.com/Shridhar2104/chat-platform/shared/models"
)

type UserRepository struct {
    db *database.PostgresDB
}

func NewUserRepository(db *database.PostgresDB) *UserRepository {
    return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(user *models.User) error {
    query := `
        INSERT INTO users (id, email, password_hash, display_name, email_verified, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `
    _, err := r.db.DB.Exec(query,
        user.ID,
        user.Email,
        user.PasswordHash,
        user.DisplayName,
        user.EmailVerified,
        user.CreatedAt,
        user.UpdatedAt,
    )
    if err != nil {
        return fmt.Errorf("failed to create user: %w", err)
    }
    return nil
}

func (r *UserRepository) GetUserByEmail(email string) (*models.User, error) {
    var user models.User
    query := `
        SELECT id, email, password_hash, display_name, avatar_url, email_verified, created_at, updated_at
        FROM users WHERE email = $1
    `
    err := r.db.DB.Get(&user, query, email)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, fmt.Errorf("user not found")
        }
        return nil, fmt.Errorf("failed to get user by email: %w", err)
    }
    return &user, nil
}

func (r *UserRepository) GetUserByID(userID uuid.UUID) (*models.User, error) {
    var user models.User
    query := `
        SELECT id, email, password_hash, display_name, avatar_url, email_verified, created_at, updated_at
        FROM users WHERE id = $1
    `
    err := r.db.DB.Get(&user, query, userID)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, fmt.Errorf("user not found")
        }
        return nil, fmt.Errorf("failed to get user by ID: %w", err)
    }
    return &user, nil
}

func (r *UserRepository) UpdateUserPassword(userID uuid.UUID, passwordHash string) error {
    query := `UPDATE users SET password_hash = $1, updated_at = $2 WHERE id = $3`
    _, err := r.db.DB.Exec(query, passwordHash, time.Now(), userID)
    if err != nil {
        return fmt.Errorf("failed to update user password: %w", err)
    }
    return nil
}

func (r *UserRepository) CreateSession(session *models.UserSession) error {
    query := `
        INSERT INTO user_sessions (id, user_id, device_id, refresh_token_hash, expires_at, created_at)
        VALUES ($1, $2, $3, $4, $5, $6)
    `
    _, err := r.db.DB.Exec(query,
        session.ID,
        session.UserID,
        session.DeviceID,
        session.RefreshTokenHash,
        session.ExpiresAt,
        session.CreatedAt,
    )
    if err != nil {
        return fmt.Errorf("failed to create session: %w", err)
    }
    return nil
}

func (r *UserRepository) GetSessionByRefreshToken(refreshTokenHash string) (*models.UserSession, error) {
    var session models.UserSession
    query := `
        SELECT id, user_id, device_id, refresh_token_hash, expires_at, created_at
        FROM user_sessions 
        WHERE refresh_token_hash = $1 AND expires_at > NOW()
    `
    err := r.db.DB.Get(&session, query, refreshTokenHash)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, fmt.Errorf("session not found or expired")
        }
        return nil, fmt.Errorf("failed to get session: %w", err)
    }
    return &session, nil
}

func (r *UserRepository) DeleteSession(sessionID uuid.UUID) error {
    query := `DELETE FROM user_sessions WHERE id = $1`
    _, err := r.db.DB.Exec(query, sessionID)
    if err != nil {
        return fmt.Errorf("failed to delete session: %w", err)
    }
    return nil
}

func (r *UserRepository) DeleteUserSessions(userID uuid.UUID, deviceID string) error {
    query := `DELETE FROM user_sessions WHERE user_id = $1 AND device_id = $2`
    _, err := r.db.DB.Exec(query, userID, deviceID)
    if err != nil {
        return fmt.Errorf("failed to delete user sessions: %w", err)
    }
    return nil
}

func (r *UserRepository) EmailExists(email string) (bool, error) {
    var count int
    query := `SELECT COUNT(*) FROM users WHERE email = $1`
    err := r.db.DB.Get(&count, query, email)
    if err != nil {
        return false, fmt.Errorf("failed to check email existence: %w", err)
    }
    return count > 0, nil
}
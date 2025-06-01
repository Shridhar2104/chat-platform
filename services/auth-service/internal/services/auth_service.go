package services

import (
    "crypto/rand"
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "time"

    "golang.org/x/crypto/bcrypt"
    "github.com/google/uuid"
    "github.com/Shridhar2104/chat-platform/shared/database"
    "github.com/Shridhar2104/chat-platform/shared/models"
    "github.com/Shridhar2104/chat-platform/auth-service/internal/repository"
)

type AuthService struct {
    userRepo   *repository.UserRepository
    jwtService *JWTService
    redis      *database.RedisClient
}

func NewAuthService(userRepo *repository.UserRepository, jwtService *JWTService, redis *database.RedisClient) *AuthService {
    return &AuthService{
        userRepo:   userRepo,
        jwtService: jwtService,
        redis:      redis,
    }
}

func (s *AuthService) Register(email, password, displayName string) (*models.User, string, string, time.Time, error) {
    // Check if email already exists
    exists, err := s.userRepo.EmailExists(email)
    if err != nil {
        return nil, "", "", time.Time{}, fmt.Errorf("failed to check email existence: %w", err)
    }
    if exists {
        return nil, "", "", time.Time{}, fmt.Errorf("email already registered")
    }

    // Hash password
    passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return nil, "", "", time.Time{}, fmt.Errorf("failed to hash password: %w", err)
    }

    // Create user
    user := &models.User{
        ID:            uuid.New(),
        Email:         email,
        PasswordHash:  string(passwordHash),
        DisplayName:   displayName,
        EmailVerified: false,
        CreatedAt:     time.Now(),
        UpdatedAt:     time.Now(),
    }

    err = s.userRepo.CreateUser(user)
    if err != nil {
        return nil, "", "", time.Time{}, fmt.Errorf("failed to create user: %w", err)
    }

    // Generate tokens
    deviceID := uuid.New().String() // Temporary device ID for registration
    accessToken, refreshToken, expiresAt, err := s.jwtService.GenerateTokenPair(user.ID, user.Email, deviceID)
    if err != nil {
        return nil, "", "", time.Time{}, fmt.Errorf("failed to generate tokens: %w", err)
    }

    return user, accessToken, refreshToken, expiresAt, nil
}

func (s *AuthService) Login(email, password, deviceID string) (*models.User, string, string, time.Time, error) {
    // Get user by email
    user, err := s.userRepo.GetUserByEmail(email)
    if err != nil {
        return nil, "", "", time.Time{}, fmt.Errorf("invalid credentials")
    }

    // Verify password
    err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
    if err != nil {
        return nil, "", "", time.Time{}, fmt.Errorf("invalid credentials")
    }

    // Generate tokens
    accessToken, refreshToken, expiresAt, err := s.jwtService.GenerateTokenPair(user.ID, user.Email, deviceID)
    if err != nil {
        return nil, "", "", time.Time{}, fmt.Errorf("failed to generate tokens: %w", err)
    }

    // Store refresh token session
    refreshTokenHash := s.hashToken(refreshToken)
    session := &models.UserSession{
        ID:               uuid.New(),
        UserID:           user.ID,
        DeviceID:         deviceID,
        RefreshTokenHash: refreshTokenHash,
        ExpiresAt:        time.Now().Add(7 * 24 * time.Hour), // 7 days
        CreatedAt:        time.Now(),
    }

    err = s.userRepo.CreateSession(session)
    if err != nil {
        return nil, "", "", time.Time{}, fmt.Errorf("failed to create session: %w", err)
    }

    return user, accessToken, refreshToken, expiresAt, nil
}

func (s *AuthService) RefreshToken(refreshToken, deviceID string) (string, string, time.Time, error) {
    // Validate refresh token
    refreshClaims, err := s.jwtService.ValidateRefreshToken(refreshToken)
    if err != nil {
        return "", "", time.Time{}, fmt.Errorf("invalid refresh token")
    }

    // Verify device ID matches
    if refreshClaims.DeviceID != deviceID {
        return "", "", time.Time{}, fmt.Errorf("device ID mismatch")
    }

    // Check if session exists in database
    refreshTokenHash := s.hashToken(refreshToken)
    session, err := s.userRepo.GetSessionByRefreshToken(refreshTokenHash)
    if err != nil {
        return "", "", time.Time{}, fmt.Errorf("session not found or expired")
    }

    // Get user details
    user, err := s.userRepo.GetUserByID(session.UserID)
    if err != nil {
        return "", "", time.Time{}, fmt.Errorf("user not found")
    }

    // Generate new token pair
    newAccessToken, newRefreshToken, expiresAt, err := s.jwtService.GenerateTokenPair(user.ID, user.Email, deviceID)
    if err != nil {
        return "", "", time.Time{}, fmt.Errorf("failed to generate new tokens: %w", err)
    }

    // Update session with new refresh token
    newRefreshTokenHash := s.hashToken(newRefreshToken)
    session.RefreshTokenHash = newRefreshTokenHash
    session.ExpiresAt = time.Now().Add(7 * 24 * time.Hour)

    // Delete old session and create new one
    s.userRepo.DeleteSession(session.ID)
    session.ID = uuid.New()
    err = s.userRepo.CreateSession(session)
    if err != nil {
        return "", "", time.Time{}, fmt.Errorf("failed to update session: %w", err)
    }

    return newAccessToken, newRefreshToken, expiresAt, nil
}

func (s *AuthService) Logout(userID uuid.UUID, deviceID string) error {
    return s.userRepo.DeleteUserSessions(userID, deviceID)
}

func (s *AuthService) GetUserByID(userID uuid.UUID) (*models.User, error) {
    return s.userRepo.GetUserByID(userID)
}

func (s *AuthService) ChangePassword(userID uuid.UUID, currentPassword, newPassword string) error {
    // Get user
    user, err := s.userRepo.GetUserByID(userID)
    if err != nil {
        return fmt.Errorf("user not found")
    }

    // Verify current password
    err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword))
    if err != nil {
        return fmt.Errorf("current password is incorrect")
    }

    // Hash new password
    newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
    if err != nil {
        return fmt.Errorf("failed to hash new password: %w", err)
    }

    // Update password
    return s.userRepo.UpdateUserPassword(userID, string(newPasswordHash))
}

func (s *AuthService) hashToken(token string) string {
    hash := sha256.Sum256([]byte(token))
    return hex.EncodeToString(hash[:])
}

func (s *AuthService) generateSecureToken() (string, error) {
    bytes := make([]byte, 32)
    _, err := rand.Read(bytes)
    if err != nil {
        return "", err
    }
    return hex.EncodeToString(bytes), nil
}
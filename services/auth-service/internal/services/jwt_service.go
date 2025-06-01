package services

import (
    "fmt"
    "time"

    "github.com/golang-jwt/jwt/v5"
    "github.com/google/uuid"
)

type JWTService struct {
    secretKey         string
    accessTokenTTL    time.Duration
    refreshTokenTTL   time.Duration
}

type Claims struct {
    UserID   uuid.UUID `json:"user_id"`
    Email    string    `json:"email"`
    DeviceID string    `json:"device_id"`
    jwt.RegisteredClaims
}

type RefreshClaims struct {
    UserID   uuid.UUID `json:"user_id"`
    DeviceID string    `json:"device_id"`
    jwt.RegisteredClaims
}

func NewJWTService(secretKey string, accessTTL, refreshTTL time.Duration) *JWTService {
    return &JWTService{
        secretKey:       secretKey,
        accessTokenTTL:  accessTTL,
        refreshTokenTTL: refreshTTL,
    }
}

func (j *JWTService) GenerateTokenPair(userID uuid.UUID, email, deviceID string) (string, string, time.Time, error) {
    // Generate access token
    now := time.Now()
    expiresAt := now.Add(j.accessTokenTTL)

    accessClaims := Claims{
        UserID:   userID,
        Email:    email,
        DeviceID: deviceID,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(expiresAt),
            IssuedAt:  jwt.NewNumericDate(now),
            NotBefore: jwt.NewNumericDate(now),
            Issuer:    "chat-platform-auth",
            Subject:   userID.String(),
        },
    }

    accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
    accessTokenString, err := accessToken.SignedString([]byte(j.secretKey))
    if err != nil {
        return "", "", time.Time{}, fmt.Errorf("failed to sign access token: %w", err)
    }

    // Generate refresh token
    refreshExpiresAt := now.Add(j.refreshTokenTTL)
    refreshClaims := RefreshClaims{
        UserID:   userID,
        DeviceID: deviceID,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(refreshExpiresAt),
            IssuedAt:  jwt.NewNumericDate(now),
            NotBefore: jwt.NewNumericDate(now),
            Issuer:    "chat-platform-auth",
            Subject:   userID.String(),
        },
    }

    refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
    refreshTokenString, err := refreshToken.SignedString([]byte(j.secretKey))
    if err != nil {
        return "", "", time.Time{}, fmt.Errorf("failed to sign refresh token: %w", err)
    }

    return accessTokenString, refreshTokenString, expiresAt, nil
}

func (j *JWTService) ValidateAccessToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return []byte(j.secretKey), nil
    })

    if err != nil {
        return nil, fmt.Errorf("failed to parse token: %w", err)
    }

    claims, ok := token.Claims.(*Claims)
    if !ok || !token.Valid {
        return nil, fmt.Errorf("invalid token")
    }

    return claims, nil
}

func (j *JWTService) ValidateRefreshToken(tokenString string) (*RefreshClaims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &RefreshClaims{}, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return []byte(j.secretKey), nil
    })

    if err != nil {
        return nil, fmt.Errorf("failed to parse refresh token: %w", err)
    }

    claims, ok := token.Claims.(*RefreshClaims)
    if !ok || !token.Valid {
        return nil, fmt.Errorf("invalid refresh token")
    }

    return claims, nil
}
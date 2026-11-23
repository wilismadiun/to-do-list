package security

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var ErrInvalidToken = errors.New("invalid token")

type AccessTokenClaims struct {
	jwt.RegisteredClaims
}

type JWTManager struct {
	secret []byte
	issuer string
	ttl    time.Duration
}

// NewJWTManager creates the JWT component for the upcoming login phase.
// The secret must contain at least 32 bytes; never hardcode it in source code.
func NewJWTManager(secret, issuer string, ttl time.Duration) (*JWTManager, error) {
	if len(secret) < 32 {
		return nil, fmt.Errorf("JWT secret must be at least 32 bytes")
	}
	if issuer == "" || ttl <= 0 {
		return nil, fmt.Errorf("JWT issuer and positive TTL are required")
	}
	return &JWTManager{secret: []byte(secret), issuer: issuer, ttl: ttl}, nil
}

func (m *JWTManager) GenerateAccessToken(userID uuid.UUID) (string, error) {
	now := time.Now()
	claims := AccessTokenClaims{RegisteredClaims: jwt.RegisteredClaims{
		Issuer: m.issuer, Subject: userID.String(),
		IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
	}}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *JWTManager) ParseAccessToken(tokenString string) (uuid.UUID, error) {
	claims := &AccessTokenClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidToken
		}
		return m.secret, nil
	}, jwt.WithIssuer(m.issuer), jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil || token == nil || !token.Valid {
		return uuid.Nil, ErrInvalidToken
	}
	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, ErrInvalidToken
	}
	return userID, nil
}

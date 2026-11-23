package unit

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"user-service/src/Applications/security"
)

func TestJWTManagerGenerateAndParse(t *testing.T) {
	manager, err := security.NewJWTManager("0123456789abcdef0123456789abcdef", "identity-service-test", time.Minute)
	require.NoError(t, err)
	userID := uuid.New()
	token, err := manager.GenerateAccessToken(userID)
	require.NoError(t, err)
	parsedID, err := manager.ParseAccessToken(token)
	require.NoError(t, err)
	require.Equal(t, userID, parsedID)
}

func TestJWTManagerRejectsInvalidSecretAndToken(t *testing.T) {
	_, err := security.NewJWTManager("short", "identity-service-test", time.Minute)
	require.Error(t, err)
	manager, err := security.NewJWTManager("0123456789abcdef0123456789abcdef", "identity-service-test", time.Minute)
	require.NoError(t, err)
	_, err = manager.ParseAccessToken("not-a-jwt")
	require.ErrorIs(t, err, security.ErrInvalidToken)
}

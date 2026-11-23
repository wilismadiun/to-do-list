package unit

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"user-service/src/Applications/security"
	"user-service/src/Applications/usecase"
	"user-service/src/Domains/entities"
	domainerr "user-service/src/Domains/errors"
	"user-service/tests/mocks"
)

type failingHasher struct{}

func (failingHasher) Hash(string) (string, error) { return "", errors.New("hash failed") }

func TestRegisterSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockUserRepository(ctrl)
	wantID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	repo.EXPECT().FindByUsername(gomock.Any(), "johndoe").Return(nil, domainerr.ErrUserNotFound)
	repo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, user *entities.User) error {
		require.Equal(t, wantID, user.ID)
		require.Equal(t, "johndoe", user.Username)
		require.NotEqual(t, "password123", user.PasswordHash)
		require.NoError(t, bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("password123")))
		return nil
	})

	uc := usecase.NewRegisterUseCaseWithIDGenerator(repo, security.NewBcryptPasswordHasher(), func() uuid.UUID { return wantID })
	got, err := uc.Execute(context.Background(), usecase.RegisterInput{Username: " johndoe ", Password: "password123"})
	require.NoError(t, err)
	require.Equal(t, wantID, got.ID)
	require.Equal(t, "johndoe", got.Username)
}

func TestRegisterValidation(t *testing.T) {
	tests := []struct {
		name  string
		input usecase.RegisterInput
	}{
		{"empty username", usecase.RegisterInput{Username: "  ", Password: "password123"}},
		{"empty password", usecase.RegisterInput{Username: "john", Password: "   "}},
		{"short password", usecase.RegisterInput{Username: "john", Password: "1234567"}},
		{"username longer than 50", usecase.RegisterInput{Username: "123456789012345678901234567890123456789012345678901", Password: "password123"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := mocks.NewMockUserRepository(ctrl)
			uc := usecase.NewRegisterUseCase(repo, security.NewBcryptPasswordHasher())
			_, err := uc.Execute(context.Background(), tt.input)
			require.ErrorIs(t, err, domainerr.ErrInvalidInput)
		})
	}
}

func TestRegisterDuplicateUsername(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockUserRepository(ctrl)
	repo.EXPECT().FindByUsername(gomock.Any(), "john").Return(&entities.User{Username: "john"}, nil)
	uc := usecase.NewRegisterUseCase(repo, security.NewBcryptPasswordHasher())
	_, err := uc.Execute(context.Background(), usecase.RegisterInput{Username: "john", Password: "password123"})
	require.ErrorIs(t, err, domainerr.ErrUserAlreadyExists)
}

func TestRegisterHashFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockUserRepository(ctrl)
	repo.EXPECT().FindByUsername(gomock.Any(), "john").Return(nil, domainerr.ErrUserNotFound)
	uc := usecase.NewRegisterUseCase(repo, failingHasher{})
	_, err := uc.Execute(context.Background(), usecase.RegisterInput{Username: "john", Password: "password123"})
	require.EqualError(t, err, "hash failed")
}

func TestRegisterCreateFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockUserRepository(ctrl)
	repo.EXPECT().FindByUsername(gomock.Any(), "john").Return(nil, domainerr.ErrUserNotFound)
	repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(errors.New("database unavailable"))
	uc := usecase.NewRegisterUseCase(repo, security.NewBcryptPasswordHasher())
	_, err := uc.Execute(context.Background(), usecase.RegisterInput{Username: "john", Password: "password123"})
	require.EqualError(t, err, "database unavailable")
}

func TestRegisterFindUsernameFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockUserRepository(ctrl)
	repo.EXPECT().FindByUsername(gomock.Any(), "john").Return(nil, errors.New("database unavailable"))
	uc := usecase.NewRegisterUseCase(repo, security.NewBcryptPasswordHasher())
	_, err := uc.Execute(context.Background(), usecase.RegisterInput{Username: "john", Password: "password123"})
	require.EqualError(t, err, "database unavailable")
}

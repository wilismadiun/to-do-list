package usecase

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"user-service/src/Applications/security"
	"user-service/src/Domains/entities"
	domainerr "user-service/src/Domains/errors"
	"user-service/src/Domains/repositories"

	"github.com/google/uuid"
)

type RegisterInput struct {
	Username string
	Password string
}

type RegisterOutput struct {
	ID       uuid.UUID
	Username string
}

type RegisterUseCase interface {
	Execute(ctx context.Context, input RegisterInput) (*RegisterOutput, error)
}

type registerUseCase struct {
	users  repositories.UserRepository
	hasher security.PasswordHasher
	newID  func() uuid.UUID
}

func NewRegisterUseCase(users repositories.UserRepository, hasher security.PasswordHasher) RegisterUseCase {
	return &registerUseCase{users: users, hasher: hasher, newID: uuid.New}
}

// NewRegisterUseCaseWithIDGenerator allows deterministic UUID generation in tests.
func NewRegisterUseCaseWithIDGenerator(users repositories.UserRepository, hasher security.PasswordHasher, newID func() uuid.UUID) RegisterUseCase {
	if newID == nil {
		newID = uuid.New
	}
	return &registerUseCase{users: users, hasher: hasher, newID: newID}
}

func (uc *registerUseCase) Execute(ctx context.Context, input RegisterInput) (*RegisterOutput, error) {
	username := strings.TrimSpace(input.Username)
	if username == "" || utf8.RuneCountInString(username) > 50 || strings.TrimSpace(input.Password) == "" || utf8.RuneCountInString(input.Password) < 8 {
		return nil, domainerr.ErrInvalidInput
	}

	existing, err := uc.users.FindByUsername(ctx, username)
	switch {
	case err == nil && existing != nil:
		return nil, domainerr.ErrUserAlreadyExists
	case err == nil && existing == nil:
		// A nil result without an error is treated as not found by this contract.
	case errors.Is(err, domainerr.ErrUserNotFound):
		// Expected: username is available.
	case err != nil:
		return nil, err
	}

	passwordHash, err := uc.hasher.Hash(input.Password)
	if err != nil {
		return nil, err
	}

	user := &entities.User{ID: uc.newID(), Username: username, PasswordHash: passwordHash}
	if err := uc.users.Create(ctx, user); err != nil {
		if errors.Is(err, domainerr.ErrUserAlreadyExists) {
			return nil, domainerr.ErrUserAlreadyExists
		}
		return nil, err
	}

	return &RegisterOutput{ID: user.ID, Username: user.Username}, nil
}

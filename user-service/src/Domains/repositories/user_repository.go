package repositories

import (
	"context"
	"user-service/src/Domains/entities"

	"github.com/google/uuid"
)

// UserRepository defines persistence operations required by the domain/application.
//
//go:generate mockgen -source=user_repository.go -destination=../../../tests/mocks/mock_user_repository.go -package=mocks
type UserRepository interface {
	Create(ctx context.Context, user *entities.User) error
	FindByUsername(ctx context.Context, username string) (*entities.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*entities.User, error)
}

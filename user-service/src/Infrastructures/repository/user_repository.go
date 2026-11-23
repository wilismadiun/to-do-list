package repository

import (
	"context"
	"errors"
	"user-service/src/Domains/entities"

	domainerr "user-service/src/Domains/errors"
	domainrepo "user-service/src/Domains/repositories"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GormUserRepository struct{ db *gorm.DB }

func NewGormUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{db: db}
}

func (r *GormUserRepository) Create(ctx context.Context, user *entities.User) error {
	err := r.db.WithContext(ctx).Create(user).Error
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return domainerr.ErrUserAlreadyExists
	}
	return err
}

func (r *GormUserRepository) FindByUsername(ctx context.Context, username string) (*entities.User, error) {
	var user entities.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domainerr.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	var user entities.User
	err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domainerr.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

var _ domainrepo.UserRepository = (*GormUserRepository)(nil)

package security

import "golang.org/x/crypto/bcrypt"

// PasswordHasher keeps password hashing independent from the use case.
type PasswordHasher interface {
	Hash(password string) (string, error)
}

type BcryptPasswordHasher struct{ Cost int }

func NewBcryptPasswordHasher() *BcryptPasswordHasher {
	return &BcryptPasswordHasher{Cost: bcrypt.DefaultCost}
}

func (h *BcryptPasswordHasher) Hash(password string) (string, error) {
	cost := h.Cost
	if cost == 0 {
		cost = bcrypt.DefaultCost
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

package errors

import "errors"

var (
	ErrUserAlreadyExists = errors.New("username already registered")
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidInput      = errors.New("invalid input")
	ErrInternalServer    = errors.New("internal server error")
)

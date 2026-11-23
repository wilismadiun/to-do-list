package handlers

import (
	"errors"
	"net/http"

	"user-service/src/Applications/dto/request"
	dtoresponse "user-service/src/Applications/dto/response"
	"user-service/src/Applications/usecase"
	domainerr "user-service/src/Domains/errors"
	commonresponse "user-service/src/commons/response"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct{ register usecase.RegisterUseCase }

func NewAuthHandler(register usecase.RegisterUseCase) *AuthHandler {
	return &AuthHandler{register: register}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req request.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		commonresponse.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	result, err := h.register.Execute(c.Request.Context(), usecase.RegisterInput{Username: req.Username, Password: req.Password})
	if err != nil {
		switch {
		case errors.Is(err, domainerr.ErrInvalidInput):
			commonresponse.Error(c, http.StatusBadRequest, "invalid username or password")
		case errors.Is(err, domainerr.ErrUserAlreadyExists):
			commonresponse.Error(c, http.StatusConflict, "username already registered")
		default:
			commonresponse.Error(c, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	commonresponse.Success(c, http.StatusCreated, "User registered successfully", dtoresponse.RegisterResponse{ID: result.ID, Username: result.Username})
}

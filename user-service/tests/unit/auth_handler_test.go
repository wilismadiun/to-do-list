package unit

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"user-service/src/Applications/usecase"
	domainerr "user-service/src/Domains/errors"
	"user-service/src/Interfaces/http/handlers"
	"user-service/tests/mocks"
)

func TestRegisterHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	id := uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb")
	tests := []struct {
		name, body  string
		status      int
		result      *usecase.RegisterOutput
		err         error
		callUseCase bool
	}{
		{"success", `{"username":"john","password":"password123"}`, http.StatusCreated, &usecase.RegisterOutput{ID: id, Username: "john"}, nil, true},
		{"invalid json", `{`, http.StatusBadRequest, nil, nil, false},
		{"missing username", `{"password":"password123"}`, http.StatusBadRequest, nil, nil, false},
		{"missing password", `{"username":"john"}`, http.StatusBadRequest, nil, nil, false},
		{"short password", `{"username":"john","password":"123"}`, http.StatusBadRequest, nil, domainerr.ErrInvalidInput, true},
		{"duplicate username", `{"username":"john","password":"password123"}`, http.StatusConflict, nil, domainerr.ErrUserAlreadyExists, true},
		{"internal error", `{"username":"john","password":"password123"}`, http.StatusInternalServerError, nil, errors.New("sensitive database detail"), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			uc := mocks.NewMockRegisterUseCase(ctrl)
			if tt.callUseCase {
				uc.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(tt.result, tt.err)
			}
			router := gin.New()
			router.POST("/register", handlers.NewAuthHandler(uc).Register)
			req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			res := httptest.NewRecorder()
			router.ServeHTTP(res, req)
			require.Equal(t, tt.status, res.Code)
			require.NotContains(t, res.Body.String(), "sensitive database detail")
			if tt.status == http.StatusCreated {
				require.Contains(t, res.Body.String(), id.String())
				require.Contains(t, res.Body.String(), "john")
				require.NotContains(t, res.Body.String(), "password_hash")
			}
		})
	}
}

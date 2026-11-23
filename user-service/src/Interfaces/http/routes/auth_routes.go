package routes

import (
	"user-service/src/Interfaces/http/handlers"

	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(api *gin.RouterGroup, handler *handlers.AuthHandler) {
	auth := api.Group("/auth")
	auth.POST("/register", handler.Register)
}

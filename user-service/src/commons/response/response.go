package response

import "github.com/gin-gonic/gin"

type Envelope struct {
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func Success(c *gin.Context, status int, message string, data any) {
	c.JSON(status, Envelope{Message: message, Data: data})
}

func Error(c *gin.Context, status int, message string) {
	c.JSON(status, Envelope{Message: message})
}

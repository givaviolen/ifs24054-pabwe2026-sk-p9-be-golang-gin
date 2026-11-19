package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Body adalah format JSON standar: {status, message, data?}.
type Body struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func Success(c *gin.Context, code int, message string, data interface{}) {
	c.JSON(code, Body{Status: "success", Message: message, Data: data})
}

func Fail(c *gin.Context, code int, message string) {
	c.AbortWithStatusJSON(code, Body{Status: "fail", Message: message})
}

func Error(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusInternalServerError, Body{Status: "error", Message: message})
}

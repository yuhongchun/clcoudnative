package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func VerifyAdmin(c *gin.Context) {
	value, _ := c.Get("role") //session is nil but there is
	if value == nil {
		c.String(http.StatusBadRequest, "you are not admin")
		c.Abort()
	}
	c.Next()
}

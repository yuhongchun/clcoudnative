package middlewares

import (
	"github.com/gin-gonic/gin"
	"go.elastic.co/apm/module/apmgin"
)

func InitMiddleware(r *gin.Engine) {
	// apm
	r.Use(gin.Logger(), apmgin.Middleware(r))
}

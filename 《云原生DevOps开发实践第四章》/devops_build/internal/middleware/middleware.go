package middleware

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"devops_build/config"
	"go.elastic.co/apm/module/apmgin"
)

func InitMiddleWare(r *gin.Engine) *gin.Engine {
	// gin.SetMode(config.ApplicationConfig.Mode)
	r.Use(apmgin.Middleware(r))
	r.Use(gin.Recovery())
	r.Use(Cors())
	r.Use(sessions.Sessions("nighting", store))
	r.Use(SetContextFromSession)
	r.Use(RequestRecord())
	if config.ApplicationConfig.Mode == "debug" {
		r.Use(gin.Logger())
	}
	return r
}

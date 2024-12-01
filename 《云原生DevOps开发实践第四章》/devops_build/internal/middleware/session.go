package middleware

import (
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/securecookie"
	"github.com/sirupsen/logrus"
)

type UserStatus struct {
	Uuid  int
	Name  string
	Admin bool
}

var store memstore.Store

func init() {
	options := sessions.Options{
		Path: "/api",
		//Domain:   "localhost",
		MaxAge:   6 * 80 * 60,
		Secure:   false,
		HttpOnly: false,
		SameSite: 1,
	}
	store = memstore.NewStore(securecookie.GenerateRandomKey(32))
	store.Options(options)
}

var UrlWriteList = []string{
	"/api/nighting-build/gitlab_callback",
	"/swagger_file/swagger.json",
}

//judge login or not
func SetContextFromSession(c *gin.Context) {
	for _, u := range UrlWriteList {
		if u == c.Request.URL.String() {
			c.Next()
			return
		}
	}
	session := sessions.Default(c)
	logrus.Info(session.ID())
	login := session.Get("login")
	if login == nil {
		if c.Request.URL.Path == "/api/nighting-build/login" {
			c.Next()
		} else if strings.Contains(c.Request.URL.Path, "swagger") || strings.Contains(c.Request.URL.Path, "swagger_file") {
			c.Next()
		} else {
			c.JSON(403, gin.H{"err_code": 403, "err_msg": "not login"})
			c.Abort()
		}
	} else {
		c.Set("role", session.Get("role")) //nil
		c.Set("uuid", session.Get("uuid"))
		c.Set("name", session.Get("name"))
		c.Next()
	}
}

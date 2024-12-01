package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"devops_build/database"
	"devops_build/database/model"
	"github.com/sirupsen/logrus"
)

type BodyReaderCloser struct {
	io.ReadCloser
	body *bytes.Buffer
}

func RequestRecord() gin.HandlerFunc {
	return func(c *gin.Context) {
		methord := c.Request.Method
		url := c.Request.URL.String()
		host := c.Request.Host
		params := ""
		query := c.Request.URL.Query()
		if len(query) != 0 {
			b, err := json.Marshal(query)
			if err != nil {
				fmt.Println("json marshal error！")
			}
			params = string(b)
		}
		body, err := ioutil.ReadAll(c.Request.Body)
		data := bytes.NewBuffer(body)
		c.Request.Body = ioutil.NopCloser(data)
		if err != nil {
			fmt.Println("read request body err!")
		}
		if len(body) != 0 {
			params = string(body)
		}
		user := model.User{}
		err = json.Unmarshal(body, &user)
		if err == nil && len(user.Account) != 0 {
			return
		}
		session := sessions.Default(c)
		uuid, ok := (session.Get("uuid")).(int)
		if !ok {
			uuid = 0
		}
		requestHistory := model.RequestHistory{
			RequestMethod: methord,
			RequestUrl:    url,
			RequestParams: params,
			Host:          host,
			UserId:        uuid,
			UpdateTime:    time.Now(),
		}
		devopsdb := database.GetDevopsDb()
		_, err = devopsdb.InsertIntoRequestHistory(c, requestHistory)
		if err != nil {
			logrus.Error(err)
		}
	}
}

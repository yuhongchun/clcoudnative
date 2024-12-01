package api

import (
	"crypto/md5"
	"encoding/hex"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"devops_build/database"
	"devops_build/database/model"
	"github.com/sirupsen/logrus"
)

// GetCurrentUser
// @Summary 获取当前用户信息
// @Description 获取当前用户信息
// @Tags user相关接口
// @Accept application/json
// @Produce application/json
// @Param name query string true "name"
// @Success 200 {object} ReturnData
// @Router /api/nighting-build/currentUser [get]
func GetCurrentUser(c *gin.Context) {
	nameI, exist := c.Get("name")
	if !exist {
		c.JSON(http.StatusBadRequest, ReturnData{
			Err_code: 0,
			Err_msg:  "name doesn't exist",
		})
		return
	}

	//duanyan
	name, ok := nameI.(string)
	if !ok {
		c.JSON(http.StatusBadRequest, ReturnData{
			Err_code: 0,
			Err_msg:  "name isn't ok for type string",
		})
		return
	}

	devopsdb := database.GetDevopsDb()
	user, err := devopsdb.GetUserByName(c, name)
	if err != nil {
		c.JSON(http.StatusBadRequest, ReturnData{
			Err_code: 0,
			Err_msg:  err.Error(),
		})
		return
	} else if user == nil {
		c.JSON(http.StatusBadRequest, ReturnData{
			Err_code: 0,
			Err_msg:  "user doesn't exist",
		})
		return
	} else {
		user.Password = "" //is ok?
		c.JSON(http.StatusOK, ReturnData{
			Err_code: 1,
			Err_msg:  "ok",
			Data:     user,
		})
		return
	}
}

// CreateUser
// @Summary 创建用户
// @Description 创建用户
// @Tags user相关接口
// @Accept application/json
// @Produce application/json
// @Param object body model.User true "用户信息"
// @Success 200 {object} ReturnData
// @Router /api/nighting-build/create_user [post]
func CreateUser(c *gin.Context) {
	createUser := model.User{}
	if err := c.BindJSON(&createUser); err != nil {
		c.JSON(http.StatusBadRequest, ReturnData{
			Err_code: 0,
			Err_msg:  err.Error(),
		})
		return
	}

	if createUser.Account == "" || createUser.Name == "" {
		c.JSON(http.StatusBadRequest, ReturnData{
			Err_code: 0,
			Err_msg:  "account and user can't be empty",
		})
		return
	}
	if createUser.Password == "" {
		c.JSON(http.StatusBadRequest, ReturnData{
			Err_code: 0,
			Err_msg:  "password can't be empty",
		})
		return
	}

	devopsdb := database.GetDevopsDb()
	user, err := devopsdb.GetUserByName(c, createUser.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, ReturnData{
			Err_code: 0,
			Err_msg:  err.Error(),
		})
		return
	} else if user != nil {
		c.JSON(http.StatusBadRequest, ReturnData{
			Err_code: 0,
			Err_msg:  "user's name has existed",
		})
		return
	}

	user, err = devopsdb.GetUserByAccount(c, createUser.Account)
	if err != nil {
		c.JSON(http.StatusBadRequest, ReturnData{
			Err_code: 0,
			Err_msg:  err.Error(),
		})
		return
	} else if user != nil {
		c.JSON(http.StatusBadRequest, ReturnData{
			Err_code: 0,
			Err_msg:  "user's account has existed",
		})
		return
	}

	sum := md5.Sum([]byte(createUser.Password))
	createUser.Password = hex.EncodeToString(sum[:])
	if user, err = devopsdb.InsertIntoUser(c, createUser); err != nil {
		c.JSON(http.StatusBadRequest, ReturnData{
			Err_code: 0,
			Err_msg:  err.Error(),
		})
		return
	} else {
		c.JSON(http.StatusOK, ReturnData{
			Err_code: 1,
			Err_msg:  "ok",
			Data:     user.Uuid,
		})
		return
	}
}

// Login 登录接口
// @Summary 登录接口
// @Description 用户登录
// @Tags user相关接口
// @Accept application/json
// @Produce application/json
// @Param object body model.User true "用户信息"
// @Success 200 {object} ReturnData
// @Router /api/nighting-build/login [post]
func Login(c *gin.Context) {
	loginuser := model.User{}
	if err := c.BindJSON(&loginuser); err != nil {
		logrus.WithContext(c).Error(err)
		c.JSON(http.StatusBadRequest, ReturnData{
			Err_code: 0,
			Err_msg:  err.Error(),
		})
		return
	}
	devopsdb := database.GetDevopsDb()
	user, err := devopsdb.GetUserByAccount(c, loginuser.Account)
	if err != nil {
		logrus.WithContext(c).Error(err)
		c.JSON(http.StatusBadRequest, ReturnData{
			Err_code: 0,
			Err_msg:  err.Error(),
		})
		return
	}
	if user == nil {
		c.JSON(http.StatusOK, ReturnData{
			Err_code: 0,
			Err_msg:  "user doesn't exist",
		})
		return
	}

	sum := md5.Sum([]byte(loginuser.Password))
	passwdstr := hex.EncodeToString(sum[:])
	if user.Password == passwdstr {
		session := sessions.Default(c)
		logrus.WithContext(c).Info(session.ID())
		session.Set("login", true)
		if user.Admin {
			session.Set("role", user.Admin)
		}
		session.Set("uuid", user.Uuid)
		session.Set("name", user.Name)
		session.Save()
		c.JSON(http.StatusOK, ReturnData{
			Err_code: 1,
			Err_msg:  "login success",
		})
		return
	} else {
		c.JSON(http.StatusBadRequest, ReturnData{
			Err_code: 0,
			Err_msg:  "incorrect password",
		})
		return
	}
}

// LoginOut
// @Summary 退出登录
// @Description 退出登录
// @Tags user相关接口
// @Accept application/json
// @Produce application/json
// @Success 200 {object} ReturnData
// @Router /api/nighting-build/loginout [get]
func LoginOut(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	c.JSON(200, gin.H{"err_msg": "ok", "err_code": 1})
}

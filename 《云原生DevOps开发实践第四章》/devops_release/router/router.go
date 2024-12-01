package router

import (
	"devops_release/api"
	"devops_release/internal/middlewares"
	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	r := gin.New()

	// 中间件注册
	middlewares.InitMiddleware(r)

	// 路由注册

	// 测试路由
	r.GET("/ping", func(context *gin.Context) {
		context.JSON(200, gin.H{
			"message": "pong",
		})
	})
	InitSysRouter(r)
	return r
}

func InitSysRouter(r *gin.Engine) *gin.RouterGroup {
	g := r.Group("/api")
	g.POST("/nighting-release/callback", api.HandleCallback)
	// 业务基础路由
	g.POST("/nighting-release/convert", api.ConvertYamlOrKV)
	g.POST("/nighting-release/add_new_tenant", api.CreateANewEnv)
	g.GET("/nighting-release/getstatus", api.GetDeployStatus)
	g.GET("/nighting-release/image_list", api.GetDockerTagList)
	g.POST("/nighting-release/compare_config", api.CompareConfigInK8sAndApollo)
	g.GET("/nighting-release/get_config", api.GetApolloConfig)
	g.POST("/nighting-release/set_config", api.SetApolloConfig)
	g.POST("/nighting-release/pub_config", api.PubApolloConfig)
	//g.GET("/nighting-release/test", api.GetStatus)
	return g
}

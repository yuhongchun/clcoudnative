package router

import (
	"net/http"

	handle "devops_build/api"
	_ "devops_build/docs"
	"devops_build/internal/middleware"
	"github.com/gin-gonic/gin"
	gs "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
)

func ProjectsRouter(r *gin.RouterGroup) {
	// list projects
	r.GET("/project", handle.ListProjects)
	r.GET("/projectname", handle.ListProjectsFuzzy)
	r.PUT("/project", handle.AddProjects)
	r.PATCH("/project", handle.PatchProjects)
	r.PUT("/project/resource/default", handle.AddNewProResourceDefault)
	r.PUT("/project/sync_gitlab", handle.SyncProjectFromGitlab)
	r.DELETE("/project", handle.DeleteProject)
	r.GET("/dingtalk_bot", handle.AddDingTalkBot)
	r.DELETE("/dingtalk_bot", handle.DeleteDingTalkBot)
	r.GET("/list_dingtalk_bot", handle.ListDingTalkBot)
	r.PUT("/dingtalk_bot", handle.AddDingTalkBot)
}

func DeploymentRouter(r *gin.RouterGroup) {
	r.GET("/selectdeployment", handle.ListDeploymentsByProIdAndNspIdAndPageSize)
	r.GET("/deployment", handle.ListDeployments)
	r.PUT("/deployment", handle.AddDeployments)
	r.PATCH("/deployment", handle.PatchDeployments)
	r.GET("/deploymentJson", handle.DeploymentJson)
	r.GET("/deploymentPatchData", handle.DeploymentPatchData)
	r.POST("/deploymentPatch", handle.DeploymentPatch)
	r.DELETE("/deployment", handle.DeleteDeployment)
	r.GET("/deployment_fuzzy", handle.ListDeploymentByFuzzy)
}

func ServiceRouter(r *gin.RouterGroup) {
	r.GET("/selectservice", handle.ListServicesByDeploymentId)
	r.GET("/service", handle.ListServices)
	r.PUT("/service", handle.AddServices)
	r.PATCH("/service", handle.PatchServices)
	r.DELETE("/service", handle.DeleteService)
}

func RouteRouter(r *gin.RouterGroup) {
	r.GET("/selectroute", handle.ListRoutesByProIdAndNspIdAndPageSize)
	r.GET("/route", handle.ListRoutes)
	r.PUT("/route", handle.AddRoutes)
	r.PATCH("/route", handle.PatchRoutes)
	r.DELETE("/route", handle.DeleteRoute)
}

func RouteUser(r *gin.RouterGroup) {
	r.POST("/login", handle.Login)
	r.GET("/currentUser", handle.GetCurrentUser)
	r.POST("/createUser", middleware.VerifyAdmin, handle.CreateUser)
	r.GET("/loginout", handle.LoginOut)
}

func ConfigRouter(r *gin.RouterGroup) {
	r.GET("/list_config_of_pro", handle.ListAllConfigOfProject)
	r.GET("/list_config_of_pro_and_nsp", handle.ListConfigOfProjectAndNamespace)
	r.GET("/list_config_of_deployment", handle.ListConfigsOfDeployment)
	r.PATCH("/config", handle.PatchConfig)
	r.DELETE("/config", handle.DeleteConfig)
	r.PUT("/config", handle.AddConfig)
}

func K8sRouter(r *gin.RouterGroup) {
	r.GET("/cluster", handle.ListCluster)
	r.GET("/namespace", handle.ListNspByClusterId)
	r.GET("/single_namespace", handle.GetNamespaceById)
}

// 跟devops-release交互的API
func NightingReleaseRoute(r *gin.RouterGroup) {
	r.GET("/release/imageList", handle.NightingReleaseImageList)

	r.POST("/release/callback", handle.NightingReleaseCallback)

	r.GET("/release/pod_info", handle.PodsInfo)
	r.POST("/release/compare_config", handle.CompareConfig)
	r.POST("/release/convert_yaml_kv", handle.ConvertYamlOrKv)
	r.GET("/release/get_config_content", handle.GetConfigContent)
	r.POST("/release/set_config", handle.SetConfigContent)
	r.POST("/release/pub_config", handle.PubConfig)
}

func InitRouter(r *gin.Engine) *gin.Engine {
	r.POST("/", func(c *gin.Context) { c.JSON(http.StatusTeapot, struct{}{}) })

	r.GET("/ping", func(c *gin.Context) {
		c.Data(http.StatusOK, "text", []byte("pong"))
	})
	r.POST("/test", func(c *gin.Context) {})
	r.GET("/swagger/*any", gs.WrapHandler(swaggerFiles.Handler))
	r.StaticFile("/swagger_file/swagger.json", "./docs/swagger.json")
	k8sapi := r.Group("/api/nighting-build")
	ProjectsRouter(k8sapi)
	DeploymentRouter(k8sapi)
	ServiceRouter(k8sapi)
	RouteRouter(k8sapi)
	K8sRouter(k8sapi)
	NightingReleaseRoute(k8sapi)
	ConfigRouter(k8sapi)

	k8sapi.POST("/gitlab_callback", handle.PipeListening)
	//暂无前端，暂时去掉用户登陆验证环节
	//userapi := r.Group("/api/nighting-build")
	//RouteUser(userapi)
	return r
}

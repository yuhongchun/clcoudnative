package api

import (
	"fmt"
	"net/http"
	"strconv"

	"devops_build/database"
	"devops_build/database/model"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ListAllConfigOfProject
// @Summary 筛选config
// @Produce json
// @Param page query string true "第几页"
// @Param size query int true "页大小"
// @Param project_id query int true "项目Id"
// @Success 200 {object} ReturnData{data=PageData}
// @Failure 400 {object} ReturnData{data=PageData}
// @Router /api/nighting-build/list_config_of_pro [get]
func ListAllConfigOfProject(c *gin.Context) {
	param := struct {
		Page      int `form:"page"`
		Size      int `form:"size"`
		ProjectId int `form:"project_id"`
	}{}
	err := c.BindQuery(&param)
	if err != nil {
		c.JSON(400, gin.H{"err_code": 0, "err_msg": "参数绑定错误！", "data": err.Error()})
		return
	}
	devopsdb := database.GetDevopsDb()
	configs, err := devopsdb.GetConfigByProjectId(c, param.ProjectId, param.Size, param.Page)
	if err != nil {
		c.JSON(500, gin.H{"err_code": 1, "err_msg": "查询错误!", "data": err.Error()})
		return
	}
	for _, config := range configs {
		d, err := devopsdb.GetDeploymentById(c, config.DeploymentId)
		if err != nil {
			continue
		}
		nsp, err := devopsdb.GetNamespaceById(c, config.NamespaceId)
		if err != nil {
			continue
		}
		config.NamespaceName = nsp.Name
		config.DeploymentName = d.DeploymentName
	}
	count, err := devopsdb.GetConfigsCountByProIdAndNspId(c, param.ProjectId, 0)
	if err != nil {
		c.JSON(500, gin.H{"err_code": 1, "err_msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ReturnData{
		Err_code: 0,
		Err_msg:  "ok",
		Data: PageData{
			Count:    count,
			ListData: configs,
		},
	})

}

// ListConfigOfProjectAndNamespace
// @Summary 筛选config
// @Produce json
// @Param page query string true "第几页"
// @Param size query int true "页大小"
// @Param project_id query int true "项目Id"
// @Param namespace_id query int true "命名空间Id"
// @Success 200 {object} ReturnData{data=PageData}
// @Failure 400 {object} ReturnData{data=PageData}
// @Router /api/nighting-build/list_config_of_pro_and_nsp [get]
func ListConfigOfProjectAndNamespace(c *gin.Context) {
	param := struct {
		Page        int `form:"page"`
		Size        int `form:"size"`
		ProjectId   int `form:"project_id"`
		NamespaceId int `form:"namespace_id"`
	}{}
	err := c.BindQuery(&param)
	if err != nil {
		c.JSON(400, gin.H{"err_code": 0, "err_msg": "参数绑定错误！", "data": err.Error()})
		return
	}
	fmt.Println(param)
	devopsdb := database.GetDevopsDb()
	configs, err := devopsdb.GetConfigByProjectIdAndNamespaceId(c, param.ProjectId, param.NamespaceId, param.Size, param.Page)
	if err != nil {
		c.JSON(500, gin.H{"err_code": 0, "err_msg": "查询错误!", "data": err.Error()})
		return
	}
	for _, config := range configs {
		d, err := devopsdb.GetDeploymentById(c, config.DeploymentId)
		if err != nil {
			continue
		}
		nsp, err := devopsdb.GetNamespaceById(c, config.NamespaceId)
		if err != nil {
			continue
		}
		config.NamespaceName = nsp.Name
		config.DeploymentName = d.DeploymentName
	}
	count, err := devopsdb.GetConfigsCountByProIdAndNspId(c, param.ProjectId, param.NamespaceId)
	if err != nil {
		c.JSON(500, gin.H{"err_code": 1, "err_msg": "查询错误!", "data": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ReturnData{
		Err_code: 0,
		Err_msg:  "ok",
		Data: PageData{
			Count:    count,
			ListData: configs,
		},
	})
}

// ListConfigOfDeployment
// @Summary 查询deployment下的config
// @Produce json
// @Param deployment_id query int true "deployment Id"
// @Success 200 {object} ReturnData{data=[]model.Config}
// @Failure 400 {object} ReturnData{data=[]model.Config}
// @Router /api/nighting-build/list_config_of_deployment [get]
func ListConfigsOfDeployment(c *gin.Context) {
	deploymentIdStr := c.Query("deployment_id")
	deploymentId, err := strconv.Atoi(deploymentIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ReturnData{Err_code: 0, Err_msg: "参数绑定错误！"})
		return
	}
	devopsdb := database.GetDevopsDb()
	configs, err := devopsdb.GetConfigByDeploymentId(c, deploymentId)
	if err != nil {
		logrus.Error(err)
		c.JSON(500, ReturnData{Err_code: 0, Err_msg: "查询错误！"})
		return
	}
	c.JSON(http.StatusOK, ReturnData{Err_code: 1, Err_msg: "ok", Data: configs})
}

// @Summary 更新某个项目发版路径下的config
// @Accept application/json
// @Param id body int true "Id"
// @Param project_id body string true "project_id"
// @Param config_name body string true "ProjetcId"
// @Param file_name body string true "file_name"
// @Param configmap_name body string true "configmap_name"
// @Param content body string true "content"
// @Param restart_after_pub body int true "restart_after_pub"
// @Param namespace_id body int  true "namespace_id"
// @Success 200 {object} ReturnData
// @Failure 400 {object} ReturnData
// @Router /api/nighting-build/config [patch]
func PatchConfig(c *gin.Context) {
	config := &model.Config{}
	err := c.BindJSON(config)
	if err != nil {
		c.JSON(400, gin.H{"err_code": 0, "err_msg": "参数绑定错误！"})
		return
	}
	devopsdb := database.GetDevopsDb()
	_, err = devopsdb.UpdateConfigById(c, *config)
	if err != nil {
		c.JSON(500, gin.H{"err_code": 0, "err_msg": "修改错误！"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"err_code": 1, "err_msg": "ok"})
}

// DeleteConfig
// @Summary 删除一个配置
// @Param id query string true "ID"
// @Success 200 {object} ReturnData
// @Failure 400 {object} ReturnData
// @Router /api/nighting-build/config [delete]
func DeleteConfig(c *gin.Context) {
	idStr := c.Query("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ReturnData{Err_code: 0, Err_msg: "参数错误！"})
		return
	}
	devopsdb := database.GetDevopsDb()
	_, err = devopsdb.DeleteConfigById(c, id)
	if err != nil {
		c.JSON(500, ReturnData{Err_code: 0, Err_msg: err.Error()})
		return
	}
	c.JSON(200, ReturnData{Err_code: 1, Err_msg: "ok"})
}

// @Summary 增加某个项目发版路径下的config
// @Accept application/json
// @Param project_id body string true "project_id"
// @Param config_name body string true "配置名"
// @Param file_name body string true "文件名"
// @Param configmap_name body string true "configmap_name"
// @Param content body string true "content"
// @Param restart_after_pub body int true "restart_after_pub"
// @Param namespace_id body int  true "namespace_id"
// @Param deployment_id body int  true "deployment_id"
// @Success 200 {object} ReturnData
// @Failure 400 {object} ReturnData
// @Router /api/nighting-build/config [put]
func AddConfig(c *gin.Context) {
	configInfo := &model.Config{}
	err := c.BindJSON(configInfo)
	if err != nil {
		c.JSON(400, ReturnData{Err_code: 0, Err_msg: "参数绑定错误！"})
		return
	}
	devopsdb := database.GetDevopsDb()
	_, err = devopsdb.InsertIntoConfig(c, *configInfo)
	if err != nil {
		c.JSON(500, ReturnData{Err_code: 0, Err_msg: err.Error()})
		return
	}

	c.JSON(200, ReturnData{Err_code: 1, Err_msg: "ok"})

}

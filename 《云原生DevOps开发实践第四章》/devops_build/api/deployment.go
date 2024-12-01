package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"devops_build/database"
	"devops_build/database/model"
	m "devops_build/database/model"
	"github.com/sirupsen/logrus"
	apiappsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"net/http"
	"sigs.k8s.io/yaml"
	"strconv"
	"strings"
	"time"
)

type ReturnData struct {
	Err_code int         `json:"err_code"`
	Err_msg  string      `json:"err_msg"`
	Data     interface{} `json:"data"`
}

type PageData struct {
	Count    int         `json:"count"`
	ListData interface{} `json:"list_data"`
}

// ListDeployments
// @Summary 列出所有项目发版路径下的deployment
// @Produce json
// @Param page query string true "第几页"
// @Param size query int true "页大小"
// @Success 200 {object} ReturnData{data=PageData}
// @Failure 400 {object} ReturnData{data=PageData}
// @Router /api/nighting-build/deployment [get]
func ListDeployments(c *gin.Context) {
	str := c.Query("page")
	s := c.Query("size")

	if str == "" || s == "" {
		c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: "Need Page and Size Param"})
		return
	}

	page, err := strconv.Atoi(str)
	if err != nil {
		c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: err.Error()})
		return
	}

	size, err := strconv.Atoi(s)
	if err != nil {
		c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: err.Error()})
		return
	}

	if page <= 0 || size <= 0 {
		c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: "Invalid Page or Size"})
		return
	}

	devopsdb := database.GetDevopsDb()
	deployments, err := devopsdb.GetDeploymentsByPageSize(c, page, size)

	if err != nil {
		c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: err.Error()})
		logrus.WithContext(c).Errorf("Error: Get Deployment failed, err: %s", err)
		return
	}
	setRealDeployments(c, deployments)
	if len(deployments) == 0 {
		c.JSON(http.StatusOK, &ReturnData{Err_code: 0, Err_msg: "Empty Data"})
		return
	}

	count, err := devopsdb.GetDeploymentCount(c)

	if err != nil {
		c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: err.Error()})
		logrus.WithContext(c).Errorf("Error: Get Deployment Count failed, err: %s", err)
		return
	}

	c.JSON(http.StatusOK, &ReturnData{Err_code: 1, Err_msg: "ok", Data: PageData{
		Count:    count,
		ListData: deployments,
	}})
}

// ListDeploymentsByProIdAndNspIdAndPageSize
// @Summary 筛选deployment
// @Produce json
// @Param page query string true "第几页"
// @Param size query int true "页大小"
// @Param proid query int true "项目Id"
// @Param nspid query int true "命名空间Id 可选 不填则只以projectid为条件"
// @Success 200 {object} ReturnData{data=PageData}
// @Failure 400 {object} ReturnData{data=PageData}
// @Router /api/nighting-build/selectdeployment [get]
func ListDeploymentsByProIdAndNspIdAndPageSize(c *gin.Context) {
	str := c.Query("page")
	s := c.Query("size")
	proidstr := c.Query("proid")
	nspidstr := c.Query("nspid")

	if str == "" || s == "" {
		c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: "Need Page and Size Param"})
		return
	}

	page, err := strconv.Atoi(str)
	if err != nil {
		c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: err.Error()})
		return
	}

	size, err := strconv.Atoi(s)
	if err != nil {
		c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: err.Error()})
		return
	}

	if page <= 0 || size <= 0 {
		c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: "Invalid Page or Size"})
		return
	}

	if proidstr == "" {
		c.JSON(http.StatusBadRequest, ReturnData{
			Err_code: 0,
			Err_msg:  "proid  can't be empty",
		})
		return
	}
	if nspidstr == "" {
		nspidstr = "0"
	}
	proid, err := strconv.Atoi(proidstr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ReturnData{
			Err_code: 0,
			Err_msg:  err.Error(),
		})
		return
	}

	nspid, err := strconv.Atoi(nspidstr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ReturnData{
			Err_code: 0,
			Err_msg:  err.Error(),
		})
		return
	}

	db := database.GetDevopsDb()
	deployments, err := db.GetDeploymentsByProIdAndNspIdAndPageSize(c, proid, nspid, page, size)
	if err != nil {
		logrus.WithContext(c).Errorf("Error: Get Deployment failed, err: %s", err)
		c.JSON(http.StatusBadRequest, &ReturnData{
			Err_code: 0,
			Err_msg:  err.Error(),
		})
		return
	}

	for _, d := range deployments {
		nsp, err := db.GetNamespaceById(c, d.NamespaceId)
		if err != nil {
			logrus.Error("err")
			continue
		}
		d.NamespaceMsg = nsp.Name + "(" + nsp.Description + ")"
	}
	setRealDeployments(c, deployments)
	count, err := db.GetDeploymentsCountByProIdAndNspId(c, proid, nspid)
	if err != nil {
		logrus.WithContext(c).Errorf("Error: Get Deployment Count failed, err: %s", err)
		c.JSON(http.StatusBadRequest, &ReturnData{
			Err_code: 0,
			Err_msg:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ReturnData{
		Err_code: 0,
		Err_msg:  "ok",
		Data: PageData{
			Count:    count,
			ListData: deployments,
		},
	})
}

// @Summary 增加某个项目发版路径下的deployment
// @Accept application/json
// @Param DeploymentName body string true "DeploymentName"
// @Param ProjetcId body int true "ProjetcId"
// @Param ChannelName body string true "ChannelName"
// @Param Content body string true "Content"
// @Param Enabled body bool true "目前都是true"
// @Param DockerRepoId body int true "DockerRepoId"
// @Param NamespaceId body int  true "NamespaceId"
// @Success 200 {object} ReturnData
// @Failure 400 {object} ReturnData
// @Router /api/nighting-build/deployment [put]
func AddDeployments(c *gin.Context) {
	deploy := m.Deployment{}
	err := c.BindJSON(&deploy)
	if err != nil {
		c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: err.Error()})
		return
	}

	devopsdb := database.GetDevopsDb()
	// exist, err := devopsdb.IfExistDeployment(c, deploy)
	// if err != nil {
	// 	nlog.WithContext(c).Errorf("Error: Judge Deployment exist failed,err: %s", err)
	// 	c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: err.Error()})
	// 	return
	// }

	// if exist {
	// 	c.JSON(http.StatusOK, &ReturnData{Err_code: 0, Err_msg: "记录已经存在，插入失败"})
	// 	return
	// }

	d, err := devopsdb.InsertIntoDeployment(c, deploy)
	if err != nil {
		c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: err.Error()})
		logrus.WithContext(c).Errorf("Error: Insert Deployment failed,err: %s", err)
		return
	}
	pro, err := devopsdb.GetProjectById(c, deploy.ProjectId)
	if err != nil {
		logrus.Error("初始化配置失败！", err)
	} else {
		cfg := model.Config{
			ProjectId:       deploy.ProjectId,
			ConfigName:      pro.ProjectName,
			FileName:        "settings.yaml",
			ConfigmapName:   pro.ProjectName,
			DeploymentId:    d.Id,
			RestartAfterPub: true,
			NamespaceId:     deploy.NamespaceId,
		}
		_, err := devopsdb.InsertIntoConfig(c, cfg)
		if err != nil {
			logrus.Error("初始化配置失败！", err)
		}
	}

	c.JSON(http.StatusOK, &ReturnData{Err_code: 0, Err_msg: "ok", Data: d.Id})
}

// @Summary 更新某个项目发版路径下的deployment
// @Accept application/json
// @Param Id body int true "Id"
// @Param DeploymentName body string true "DeploymentName"
// @Param ProjetcId body int true "ProjetcId"
// @Param ChannelName body string true "ChannelName"
// @Param Content body string true "Content 可以是yaml也可以是json字符串"
// @Param Enabled body bool true "目前都是true"
// @Param DockerRepoId body int true "DockerRepoId"
// @Param NamespaceId body int  true "NamespaceId"
// @Success 200 {object} ReturnData
// @Failure 400 {object} ReturnData
// @Router /api/nighting-build/deployment [patch]
func PatchDeployments(c *gin.Context) {
	deploy := m.Deployment{}
	err := c.BindJSON(&deploy)
	if err != nil {
		c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: err.Error()})
		return
	}
	deploymentStruct := &apiappsv1.Deployment{}
	err = yaml.Unmarshal([]byte(deploy.Content), deploymentStruct)
	if err != nil {
		err = json.Unmarshal([]byte(deploy.Content), deploymentStruct)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"err_code": 0, "err_msg": "deployment内容格式错误！"})
			return
		}
		b, err := yaml.Marshal(deploymentStruct)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"err_code": 0, "err_msg": "deployment内容格式错误！"})
		}
		deploy.Content = string(b)
	}
	devopsdb := database.GetDevopsDb()
	changeBeforeDeployment, err := devopsdb.GetDeploymentById(c, deploy.Id)
	if err != nil {
		logrus.Error(err)
	}
	result, err := devopsdb.UpdateDeployment(c, deploy)
	if err != nil {
		c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: err.Error()})
		logrus.WithContext(c).Errorf("Patch Deployment failed err:", err)
		return
	}
	if result {
		c.JSON(http.StatusOK, &ReturnData{Err_code: 0, Err_msg: "ok"})
	} else {
		c.JSON(http.StatusOK, &ReturnData{Err_code: 0, Err_msg: "update failed"})
		return
	}
	session := sessions.Default(c)
	uuid, ok := session.Get("uuid").(int)
	if !ok {
		uuid = 0
	}
	changeAfterDeployment := deploy
	opsHistory := model.OpsHistory{
		ResourceType: "deployment",
		OpsType:      "update_deployment",
		ResourceId:   deploy.Id,
		UserId:       uuid,
		UpdateTime:   time.Now(),
		ChangeBefore: changeBeforeDeployment.Content,
		ChangeAfter:  changeAfterDeployment.Content,
	}
	if strings.TrimSpace(opsHistory.ChangeBefore) != strings.TrimSpace(opsHistory.ChangeAfter) {
		_, err = devopsdb.InsertIntoOpsHistory(c, opsHistory)
		if err != nil {
			logrus.Error(err)
		}
	}
}

// @Summary 获取deployment的结构体
// @Accept application/json
// @Param deployment_id query int true "deployment_id"
// @Param type query string false "type json或者yaml决定着返回数据的格式"
// @Success 200 {object} ReturnData
// @Failure 400 {object} ReturnData
// @Router /api/nighting-build/deploymentJson [get]
func DeploymentJson(c *gin.Context) {
	idS := c.Request.URL.Query().Get("deployment_id")
	type1 := c.Request.URL.Query().Get("type")
	id, err := strconv.Atoi(idS)
	if err != nil {
		c.JSON(400, gin.H{
			"err_code": 0,
			"err_msg":  "参数错误！",
		})
		return
	}
	if id == 0 {
		c.JSON(400, gin.H{
			"err_code": 0,
			"err_msg":  "参数错误！",
		})
		return
	}
	devopsdb := database.GetDevopsDb()
	d, err := devopsdb.GetDeploymentById(c, id)
	if err != nil {
		c.JSON(200, gin.H{
			"err_code": 0,
			"err_msg":  "该deployment不存在",
		})
		return
	}
	setRealDeployment(c, d)
	if type1 == "yaml" {
		c.JSON(http.StatusOK, gin.H{"err_code": 1, "err_msg": "ok", "data": d.Content})
		return
	}

	deploymentStruct := &apiappsv1.Deployment{}
	err = yaml.Unmarshal([]byte(d.Content), deploymentStruct)
	if err != nil {
		logrus.WithContext(c).Error(err)
		c.JSON(200, gin.H{"err_code": 0, "err_msg": "yaml解析出错！"})
		return
	}
	c.JSON(200, gin.H{"err_code": 1, "err_msg": "ok", "data": deploymentStruct})
}

// @Summary 面板或yaml更新deployment content
// @Accept application/json
// @Param object body SimpleDeployment true "面板deployment数据"
//
// @Success 200 {object} ReturnData
// @Failure 400 {object} ReturnData
// @Router /api/nighting-build/deploymentPatch [post]
func DeploymentPatch(c *gin.Context) {
	simpleDeployment := SimpleDeployment{}
	err := c.BindJSON(&simpleDeployment)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err_code": 0, "err_msg": "参数绑定错误"})
		return
	}
	devopsdb := database.GetDevopsDb()
	d, err := devopsdb.GetDeploymentById(c, simpleDeployment.Id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err_code": 0, "err_msg": "deployment不存在！"})
		return
	}

	deploymentStruct := &apiappsv1.Deployment{}
	if len(strings.TrimSpace(simpleDeployment.Content)) != 0 {
		d.Content = simpleDeployment.Content
		err = yaml.Unmarshal([]byte(d.Content), &deploymentStruct)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"err_code": 0, "err_msg": "deployment yaml 不合法！"})
			return
		}
		_, err = devopsdb.UpdateDeployment(c, *d)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"err_code": 0, "err_msg": "数据更新错误！"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"err_code": 1, "err_msg": "ok"})
		return
	}
	err = yaml.Unmarshal([]byte(d.Content), &deploymentStruct)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err_code": 0, "err_msg": "deployment yaml 不合法！"})
		return
	}

	if len(deploymentStruct.Spec.Template.Spec.Containers) != len(simpleDeployment.Containers) {
		c.JSON(http.StatusOK, gin.H{"err_code": 0, "err_msg": "容器数量不等！"})
		return
	}
	for i, c := range deploymentStruct.Spec.Template.Spec.Containers {
		deploymentStruct.Spec.Template.Spec.Containers[i].Image = simpleDeployment.Containers[i].Image
		c.Env = []corev1.EnvVar{}
		for k, v := range simpleDeployment.Containers[i].Env {
			c.Env = append(c.Env, corev1.EnvVar{
				Name:  k,
				Value: v,
			})
		}
		deploymentStruct.Spec.Template.Spec.Containers[i].Env = c.Env
	}
	fmt.Println(deploymentStruct.Spec.Template.Spec.Containers[0].Env)
	y, err := yaml.Marshal(deploymentStruct)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"err_code": 0, "err_msg": "数据格式错误！"})
		return
	}
	d.Content = string(y)
	_, err = devopsdb.UpdateDeployment(c, *d)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"err_code": 0, "err_msg": "数据更新错误！"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"err_code": 1, "err_msg": "ok"})
}

// @Summary 获取deployment面板数据的接口
// @Accept application/json
// @Param deployment_id query int true "deployment_id"
// @Success 200 {object} ReturnData
// @Failure 400 {object} ReturnData
// @Router /api/nighting-build/deploymentPatchData [get]
func DeploymentPatchData(c *gin.Context) {
	id, err := strconv.Atoi(c.Request.URL.Query().Get("deployment_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err_code": 0, "err_msg": "参数绑定错误"})
		return
	}

	devopsdb := database.GetDevopsDb()
	d, err := devopsdb.GetDeploymentById(c, id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"err_code": 0, "err_msg": "deployment 不存在"})
		return
	}
	setRealDeployment(c, d)
	simpledeploy := SimpleDeployment{}
	deploymentStruct := apiappsv1.Deployment{}
	err = yaml.Unmarshal([]byte(d.Content), &deploymentStruct)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err_code": 0, "err_msg": "deployment yaml 不合法！"})
		return
	}

	simpledeploy.Id = d.Id
	for _, c := range deploymentStruct.Spec.Template.Spec.Containers {
		env := map[string]string{}
		for _, e := range c.Env {
			env[e.Name] = e.Value
		}
		simpledeploy.Containers = append(simpledeploy.Containers, SimpleContainer{
			Image: c.Image,
			Env:   env,
		})
	}
	c.JSON(http.StatusOK, gin.H{"err_code": 1, "err_msg": "ok", "data": simpledeploy})
}

type DeploymentPatcher struct {
	apiappsv1.Deployment
}

type SimpleDeployment struct {
	Id         int               `json:"id"`
	Containers []SimpleContainer `json:"containers"`
	Content    string            `json:"content"`
}
type SimpleContainer struct {
	Image string            `json:"image"`
	Env   map[string]string `json:"env"`
}

// DeleteDeployment
// @Summary 删除一个项目
// @Param deployid query string true "ID"
// @Success 200 {object} ReturnData
// @Failure 400 {object} ReturnData
// @Router /api/nighting-build/deployment [delete]
func DeleteDeployment(c *gin.Context) {
	strid := c.Query("deployid")
	deployid, err := strconv.Atoi(strid)
	if err != nil {
		c.JSON(http.StatusBadRequest, &ReturnData{
			Err_code: 0,
			Err_msg:  err.Error(),
		})
		return
	}
	devopsdb := database.GetDevopsDb()
	_, err = devopsdb.DeleteDeploymentById(c, deployid)
	if err != nil {
		c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: err.Error()})
		logrus.WithContext(c).Errorf("Delete Deployment failed err:", err)
		return
	}

	//删除该deployment 下的所有资源

	c.JSON(http.StatusOK, &ReturnData{
		Err_code: 1,
		Err_msg:  "ok",
	})
}

// ListDeploymentByFuzzyFind
// @Summary 模糊查询得出所需的deployment
// @Produce json
// @Param fuzzystr query string true "输入的项目查询字符串"
// @Param project_id query int true "必须"
// @Param namespace_id query int false "传的话则加上该筛选条件"
// @Success 200 {object} ReturnData
// @Router /api/nighting-build/deployment_fuzzy [get]
func ListDeploymentByFuzzy(c *gin.Context) {
	fuzzstr := c.Query("fuzzystr")
	if fuzzstr == "" {
		c.JSON(http.StatusBadRequest, &ReturnData{Err_code: 0, Err_msg: "Need fuzzystr Param"})
		return
	}

	projectStr := c.Query("project_id")
	projectId, err := strconv.Atoi(projectStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ReturnData{Err_code: 0, Err_msg: "projectId参数类型错误"})
		return
	}
	namespaceStr := c.Query("namespace_id")
	namespaceId, err := strconv.Atoi(namespaceStr)
	if err != nil {
		namespaceId = 0
	}

	devopsdb := database.GetDevopsDb()
	res, err := devopsdb.GetDeploymentByFuzzyFind(c, fuzzstr, namespaceId, projectId)
	if err != nil {
		c.JSON(500, gin.H{"err_code": 0, "err_msg": err.Error()})
		return
	}
	if len(res) == 0 {
		c.JSON(200, gin.H{"err_code": 1, "err_msg": "ok", "list_data": []model.Deployment{}})
		return
	}
	c.JSON(200, gin.H{"err_code": 1, "err_msg": "ok", "list_data": res})
}

func setRealDeployments(ctx context.Context, deployments []*model.Deployment) {
	for _, d := range deployments {
		if d != nil {
			setRealDeployment(ctx, d)
		}
	}
}

func setRealDeployment(ctx context.Context, deployment *model.Deployment) {
	devopsdb := database.GetDevopsDb()
	pro, err := devopsdb.GetProjectById(ctx, deployment.ProjectId)
	if err != nil {
		logrus.Error()
		return
	}
	nsp, err := devopsdb.GetNamespaceById(ctx, deployment.NamespaceId)
	if err != nil {
		logrus.Error()
		return
	}
	deployment.Content = strings.ReplaceAll(deployment.Content, "${__project_name}", pro.ProjectName)
	deployment.Content = strings.ReplaceAll(deployment.Content, "${__k8s_namespace}", nsp.Name)
}

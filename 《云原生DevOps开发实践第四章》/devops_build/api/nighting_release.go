package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	nightingrelease "devops_build/controller/nighting_release"
	statuscode "devops_build/util/status_code"
	"github.com/sirupsen/logrus"
)

// Callback nighting-release回调接口
// @Summary nighting-release回调接口
// @Description 重启  部署  patch_image event_type为build|restart|patch_image
// @Tags 发版相关接口
// @Accept application/json
// @Produce application/json
// @Param object body nightingrelease.ReleaseInfo true "主要是获取event_type字段等"
// @Success 200 {object} nightingrelease.OpsInfo
// @Router /api/nighting-build/release/callback [post]
func NightingReleaseCallback(c *gin.Context) {
	// TODO: 操作通知，给出操作人，操作对象，操作结果发送通知到钉钉
	releaseInfo := nightingrelease.ReleaseInfo{}
	err := c.BindJSON(&releaseInfo)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err_code": statuscode.PARAMS_BIND_ERR, "err_msg": statuscode.ErrMsg[statuscode.PARAMS_BIND_ERR]})
	}
	body, err := nightingrelease.CallBack(c, releaseInfo)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"err_code": 0, "err_msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, body)
}

// Callback nighting-release获取镜像接口
// @Summary nighting-release获取镜像列表接口
// @Description 获取镜像列表
// @Tags 发版相关接口
// @Accept application/json
// @Produce application/json
// @Param repo query string true "获取镜像"
// @Success 200 {object} ReleaseImageRes
// @Router /api/nighting-build/release/imageList [get]
func NightingReleaseImageList(c *gin.Context) {
	repo := c.Query("repo")
	body, err := nightingrelease.ImageList(c, repo)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"err_code": 0, "err_msg": "查询错误！"})
	}
	c.JSON(http.StatusOK, body)
}

// Callback nighting-release获取镜像接口
// @Summary nighting-release获取镜像列表接口
// @Description 获取镜像列表
// @Tags 发版相关接口
// @Accept application/json
// @Produce application/json
// @Param project_id query int true "项目id"
// @Param namespace_id query int true "namespaceid"
// @Param deployment_name query string true "deployment name"
// @Success 200 {object} ReleasePodsRes
// @Router /api/nighting-build/release/pod_info [get]
func PodsInfo(c *gin.Context) {
	proId, err := strconv.Atoi(c.Query("project_id"))
	nspId, err := strconv.Atoi(c.Query("namespace_id"))
	depName := c.Query("deployment_name")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err_code": 0, "err_msg": "参数绑定错误！"})
		return
	}
	podInfos, err := nightingrelease.Pods(c, proId, nspId, depName)
	if err != nil {
		c.JSON(200, gin.H{"err_code": 0, "err_msg": "获取podinfo出错！"})
		logrus.Error(err)
		return
	}
	c.JSON(200, gin.H{"err_code": 1, "err_msg": "ok", "data": podInfos})

}

// CompareConfig 比较配置接口
// @Summary nighting-release比较配置接口
// @Description 比较配置中心配置和k8s中配置的一致性
// @Tags nighting-release
// @Accept application/json
// @Produce application/json
// @Param object body nightingrelease.CompareInfo true "比较配置信息"
// @Success 200 {object} ReleaseCompareConfig
// @Router /api/nighting-build/release/compare_config [post]
func CompareConfig(c *gin.Context) {
	compareInfo := struct {
		ConfigId int `json:"config_id"`
	}{}
	err := c.BindJSON(&compareInfo)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err_code": 0, "err_msg": "参数绑定错误！"})
		return
	}

	resData, err := nightingrelease.CompareConfig(c, nightingrelease.CompareInfo{
		ConfigId: compareInfo.ConfigId,
	})
	if err != nil {
		c.JSON(500, gin.H{"err_code": 0, "err_msg": "调用nighting-release失败"})
		return
	}
	res := ReleaseCompareConfig{}
	err = json.Unmarshal(resData, &res)
	if err != nil {
		c.JSON(500, gin.H{"err_code": 0, "err_msg": err.Error()})
		return
	}
	c.JSON(200, res)
}

// ConvertYamlOrKv kv<----->yaml转换接口
// @Summary nighting-release yaml转换接口
// @Description yaml和kv互相转换
// @Tags nighting-release
// @Accept application/json
// @Produce application/json
// @Param object body nightingrelease.ConvertData true "转换信息 class为 yaml或者kv"
// @Success 200 {object} ReleaseConvertRes
// @Router /api/nighting-build/release/convert_yaml_kv [post]
func ConvertYamlOrKv(c *gin.Context) {
	convertData := nightingrelease.ConvertData{}
	err := c.BindJSON(&convertData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err_code": 0, "err_msg": "参数绑定错误！"})
		return
	}
	data, err := nightingrelease.ConvertYamlOrKv(convertData)
	if err != nil {
		c.JSON(500, gin.H{"err_code": 0, "err_msg": err.Error()})
		return
	}
	res := ReleaseConvertRes{}
	err = json.Unmarshal(data, &res)
	if err != nil {
		c.JSON(500, gin.H{"err_code": 0, "err_msg": err.Error()})
		return
	}
	c.JSON(200, res)
}

// GetConfigContent
// @Summary 筛选config
// @Produce json
// @Param config_id query int true "配置信息Id"
// @Success 200 {object} ReleaseGetConfigRes
// @Router /api/nighting-build/release/get_config_content [get]
func GetConfigContent(c *gin.Context) {
	configIdStr := c.Query("config_id")
	configId, err := strconv.Atoi(configIdStr)
	if err != nil {
		c.JSON(400, gin.H{"err_code": 0, "err_msg": err.Error()})
		return
	}
	data, err := nightingrelease.GetConfig(configId)
	if err != nil {
		c.JSON(500, gin.H{"err_code": 0, "err_msg": err.Error()})
		return
	}
	res := ReleaseGetConfigRes{}
	err = json.Unmarshal(data, &res)
	if err != nil {
		c.JSON(400, gin.H{"err_code": 0, "err_msg": err.Error()})
		return
	}
	c.JSON(200, res)

}

// SetConfigContent 更新配置接口
// @Summary nighting-release 更新配置接口
// @Description 更新配置内容
// @Tags nighting-release
// @Accept application/json
// @Produce application/json
// @Param object body nightingrelease.SetConfigInfo true "保存配置信息 file_type为 yaml或者kv"
// @Success 200 {object} ReturnData
// @Router /api/nighting-build/release/set_config [post]
func SetConfigContent(c *gin.Context) {
	setConfigInfo := &nightingrelease.SetConfigInfo{}
	err := c.BindJSON(setConfigInfo)
	if err != nil {
		c.JSON(400, ReturnData{Err_code: 0, Err_msg: err.Error()})
		return
	}
	data, err := nightingrelease.SetConfig(*setConfigInfo)
	if err != nil {
		c.JSON(500, ReturnData{Err_code: 0, Err_msg: err.Error()})
		return
	}
	res := ReturnData{}
	err = json.Unmarshal(data, &res)
	if err != nil {
		c.JSON(500, ReturnData{Err_code: 0, Err_msg: err.Error()})
		return
	}
	c.JSON(200, res)
}

// PubConfig 发布配置
// @Summary nighting-release 发布配置接口
// @Description 发布更新的配置内容
// @Tags nighting-release
// @Accept application/json
// @Produce application/json
// @Param config_id query int true "配置信息Id"
// @Success 200 {object} ReturnData
// @Router /api/nighting-build/release/pub_config [post]
func PubConfig(c *gin.Context) {
	configIdstr := c.Query("config_id")
	configId, err := strconv.Atoi(configIdstr)
	if err != nil {
		c.JSON(400, ReturnData{Err_code: 0, Err_msg: "参数错误"})
		return
	}
	res, err := nightingrelease.PubConfig(configId)
	if err != nil {
		c.JSON(500, ReturnData{Err_code: 0, Err_msg: err.Error()})
		return
	}
	resMsg := &ResponseMsg{}
	err = json.Unmarshal(res, resMsg)
	if err != nil {
		c.JSON(500, ReturnData{Err_code: 0, Err_msg: err.Error()})
		return
	}
	c.JSON(200, resMsg)
}

type ReleaseGetConfigRes struct {
	ResponseMsg
	Content string `json:"content"`
}

type ReleaseConvertRes struct {
	ResponseMsg
	Data string `json:data`
}

type ReleaseCompareConfig struct {
	ResponseMsg
	CompareRes nightingrelease.CompareConfigRes `json:"compare_res"`
}
type ReleaseImageRes struct {
	ResponseMsg
	ImageList []string `json:"image_list"`
}
type ReleaseCallBackRes struct {
	ResponseMsg
	OpsInfo nightingrelease.OpsInfo `json:"ops_info"`
}
type ReleasePodsRes struct {
	ResponseMsg
	Data nightingrelease.DeploymentPodInfo `json:"data"`
}
type ResponseMsg struct {
	ErrCode int    `json:"err_code"`
	ErrMsg  string `json:"err_msg"`
}

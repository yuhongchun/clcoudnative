package nightingrelease

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"devops_build/config"
	"devops_build/database"
	"devops_build/database/model"
	httputil "devops_build/util/http_util"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type ReleaseInfo struct {
	Id          string `json:"id"`
	EventType   string `json:"event_type"`
	TagId       string `json:"tag_id"`
	Deployments []int  `json:"deployments"`
	ProjectId   int    `json:"project_id"`
	NamespaceId int    `json:"namespace_id"`
	OpsUser     string `json:"ops_user"`
}
type ReleaseRes struct {
}
type OpsInfo struct {
	ProjectName   string         `json:"project_name"`
	Type          string         `json:"type"`
	OpsProResults []OpsProResult `json:"ops_pro_results"`
	Channel       string         `json:"channel"`
	Status        string         `json:"status"`
}
type OpsProResult struct {
	Namespace      string   `json:"namespace"`
	ClusterName    string   `json:"cluster_name"`
	Status         string   `json:"status"`
	Message        string   `json:"message"`
	DeploymentErrs []string `json:"deployment_errs"`
	ServiceErrs    []string `json:"service_errs"`
	ConfigMapErrs  []string `json:"configmap_errs"`
}
type CompareInfo struct {
	DeploymentId int    `json:"deployment_id"`
	Cluster      string `json:"cluster"`
	Namespace    string `json:"namespace"`
	ConfigId     int    `json:"config_id"`
}
type ConvertData struct {
	Content string `json:"content"`
	Class   string `json:"class"`
}
type SetConfigInfo struct {
	ConfigId int    `json:"config_id"`
	Content  string `json:"content"`
	FileType string `json:"file_type"`
}

func CallBack(ctx *gin.Context, releaseInfo ReleaseInfo) (*CallbackReturnData, error) {
	data, err := json.Marshal(releaseInfo)
	if err != nil {
		logrus.Error("json编码错误！err:", err)
		return nil, err
	}
	if len(releaseInfo.Deployments) == 0 {
		return nil, fmt.Errorf("deployments is empty!")
	}
	devopsDb := database.GetDevopsDb()
	changeBeforeDeployments := []model.Deployment{}
	for _, d := range releaseInfo.Deployments {
		deployment, err := devopsDb.GetDeploymentById(ctx, d)
		if err == nil {
			changeBeforeDeployments = append(changeBeforeDeployments, *deployment)
		} else {
			return nil, err
		}
	}
	body, err := httputil.SendHttpRequest("POST", map[string]string{"Content-type": "application/json"}, config.NightingHostConfig.NightingReleaseHost+"/api/nighting-release/callback", data)
	if err != nil {
		logrus.Error("发送请求错误！err:", err)
		return nil, err
	}
	session := sessions.Default(ctx)
	userId, ok := session.Get("uuid").(int)
	if !ok {
		userId = 0
	}

	for i, d := range releaseInfo.Deployments {
		deployment, err := devopsDb.GetDeploymentById(ctx, d)
		if err != nil {
			logrus.Error(err)
			break
		}
		opsHistory := model.OpsHistory{
			ResourceType: "deployment",
			OpsType:      releaseInfo.EventType,
			ResourceId:   d,
			UserId:       userId,
			UpdateTime:   time.Now(),
			ChangeBefore: changeBeforeDeployments[i].Content,
			ChangeAfter:  deployment.Content,
		}
		if strings.TrimSpace(opsHistory.ChangeBefore) != strings.TrimSpace(opsHistory.ChangeAfter) {
			_, err = devopsDb.InsertIntoOpsHistory(ctx, opsHistory)
			if err != nil {
				logrus.Error(err)
			}
		}
	}
	returnData := CallbackReturnData{}
	err = json.Unmarshal(body, &returnData)
	if err != nil {
		return nil, err
	}
	return &returnData, nil
}

type CallbackReturnData struct {
	Err_code int         `json:"err_code"`
	Err_msg  string      `json:"err_msg"`
	OpsInfo  interface{} `json:"ops_info"`
}
type ImagelistReturnData struct {
	Err_code  int      `json:"err_code"`
	Err_msg   string   `json:"err_msg"`
	ImageList []string `json:"image_list"`
}

func ImageList(ctx context.Context, repo string) (*ImagelistReturnData, error) {
	body, err := httputil.SendHttpRequest("GET", map[string]string{"Content-type": "application/json"}, config.NightingHostConfig.NightingReleaseHost+"/api/nighting-release/image_list?repo="+repo, nil)
	if err != nil {
		logrus.Error("发送请求错误！", err)
		return nil, err
	}
	returnData := &ImagelistReturnData{}
	err = json.Unmarshal(body, returnData)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	return returnData, nil
}

func Pods(ctx context.Context, projectId int, nspId int, deploymentName string) (*DeploymentPodInfo, error) {
	devopsdb := database.GetDevopsDb()
	nsp, err := devopsdb.GetNamespaceById(ctx, nspId)
	if err != nil {
		return nil, err
	}
	url := config.NightingHostConfig.NightingReleaseHost + "/api/nighting-release/getstatus?nsname=" + nsp.Name + "&dename=" + deploymentName
	body, err := httputil.SendHttpRequest("GET", nil, url, nil)
	if err != nil {
		logrus.Error("发送请求错误！")
		return nil, err
	}
	podInfos := []PodInfo{}
	err = json.Unmarshal(body, &podInfos)
	if err != nil {
		logrus.Error("json解析错误！err:", err)
		return nil, err
	}

	return &DeploymentPodInfo{
		Namespace: nsp.Name,
		PodInfos:  podInfos,
	}, nil
}

func CompareConfig(ctx context.Context, compareInfo CompareInfo) ([]byte, error) {
	url := config.NightingHostConfig.NightingReleaseHost + "/api/nighting-release/compare_config"
	headers := map[string]string{
		"Content-type": "application/json",
	}
	data, err := json.Marshal(compareInfo)
	if err != nil {
		return nil, err
	}
	body, err := httputil.SendHttpRequest("POST", headers, url, data)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func ConvertYamlOrKv(convertData ConvertData) ([]byte, error) {
	url := config.NightingHostConfig.NightingReleaseHost + "/api/nighting-release/convert"
	headers := map[string]string{
		"Content-type": "application/json",
	}
	data, err := json.Marshal(convertData)
	if err != nil {
		return nil, err
	}
	body, err := httputil.SendHttpRequest("POST", headers, url, data)
	if err != nil {
		return nil, fmt.Errorf("请求nighting-release错误！")
	}
	return body, nil

}

func GetConfig(configId int) ([]byte, error) {
	url := fmt.Sprintf("%s%s%d", config.NightingHostConfig.NightingReleaseHost, "/api/nighting-release/get_config?config_id=", configId)
	body, err := httputil.SendHttpRequest("GET", nil, url, nil)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func SetConfig(setConfigInfo SetConfigInfo) ([]byte, error) {
	url := fmt.Sprintf("%s%s", config.NightingHostConfig.NightingReleaseHost, "/api/nighting-release/set_config")
	headers := map[string]string{
		"Content-type": "application/json",
	}
	data, err := json.Marshal(&setConfigInfo)
	if err != nil {
		return nil, err
	}
	body, err := httputil.SendHttpRequest("POST", headers, url, data)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func PubConfig(configId int) ([]byte, error) {
	url := fmt.Sprintf("%s%s%d", config.NightingHostConfig.NightingReleaseHost, "/api/nighting-release/pub_config?config_id=", configId)
	body, err := httputil.SendHttpRequest("POST", nil, url, nil)
	if err != nil {
		return nil, err
	}
	return body, nil
}

type CompareConfigRes struct {
	Config1 map[string]string `json:"apollo_config"`
	Config2 map[string]string `json:"k8s_config"`
	IsEqual bool              `json:"is_equal"`
}
type PodInfo struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	RunTime string `json:"run_time"`
}
type DeploymentPodInfo struct {
	Namespace string    `json:"namespace"`
	PodInfos  []PodInfo `json:"podinfos"`
}

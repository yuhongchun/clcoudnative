package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	k8sutil "devops_release/util/k8s_util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	//clientset "k8s.io/client-go/kubernetes"
	"log"
	//"gitlab.yiban.io/we-work-go-team/nighting-release/internal/build"
	"devops_release/database"
	"devops_release/internal/buildv2"
	addnewtenant "devops_release/internal/service/add_new_tenant"
	"devops_release/internal/service/compare"
	k8sresource "devops_release/internal/service/k8s_resource"
	"devops_release/util/apollo"
	"devops_release/util/model"
	"devops_release/util/myyaml"
	nlog "github.com/sirupsen/logrus"
)

func HandleCallback(c *gin.Context) {
	var releaseInfo buildv2.ReleaseInfo
	err := c.ShouldBindJSON(&releaseInfo)
	if err != nil {
		nlog.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{"err_msg": "参数绑定错误！"})
		return
	}
	nlog.Info("releaseInfo:", releaseInfo)
	opsInfo, err := buildv2.BuildProject(c.Request.Context(), releaseInfo)
	if err != nil {
		c.JSON(500, gin.H{"err_code": 0, "err_msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{"err_code": 1, "err_msg": "ok", "ops_info": opsInfo})
}

func GetDockerTagList(c *gin.Context) {
	repo := c.Request.URL.Query().Get("repo")
	res, err := k8sresource.GetDockerTagList(repo)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"err_code": 0, "err_msg": "获取镜像失败！"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"err_code": 1, "err_msg": "查询成功！", "image_list": res})
}

func ConvertYamlOrKV(c *gin.Context) {
	convertData := &struct {
		Content string `json:"content"`
		Class   string `json:"class"`
	}{}
	err := c.BindJSON(convertData)
	if err != nil {
		c.JSON(500, gin.H{"err_code": 1, "err_msg": "参数解析错误！"})
		return
	}
	content := ""
	if convertData.Class == "yaml" {
		kvs, err := apollo.YamlTransitionApollo(convertData.Content)
		if err != nil {
			c.JSON(500, gin.H{"err_code": 1, "err_msg": "yaml格式有误！！"})
			return
		}

		for k, v := range kvs {

			vString, ok := v.(string)
			if ok {
				isnumber := IsNumber(v.(string))
				if isnumber {
					vString = `"` + v.(string) + `"`
				} else {
					vString = SetContainerSpecialCh(v.(string))
				}
				content = fmt.Sprintf("%s%s = %s\n ", content, k, vString)

			}
			vInt, ok := v.(int)
			if ok {
				content = fmt.Sprintf("%s%s = %d\n ", content, k, vInt)
			}
			vFloat, ok := v.(float64)
			if ok {
				content = fmt.Sprintf("%s%s = %f\n ", content, k, vFloat)
			}
			vFloat32, ok := v.(float32)
			if ok {
				content = fmt.Sprintf("%s%s = %f\n ", content, k, vFloat32)
			}
			vBool, ok := v.(bool)
			if ok {
				content = fmt.Sprintf("%s%s = %t\n ", content, k, vBool)

			}
		}
	} else if convertData.Class == "kv" {
		items := []model.Item{}
		lines := strings.Split(convertData.Content, "\n")
		for _, line := range lines {
			lineSplit := strings.Split(line, "=")
			key := ""
			value := ""
			if len(lineSplit) != 2 {
				key = strings.TrimSpace(lineSplit[0])
			} else {
				key = strings.TrimSpace(lineSplit[0])
				value = strings.TrimSpace(lineSplit[1])
			}
			item := model.Item{
				Key:   key,
				Value: value,
			}
			items = append(items, item)
		}
		myyaml := myyaml.NewYaml(items)
		content = myyaml.ToString()
	}
	c.JSON(http.StatusOK, gin.H{"err_code": 1, "err_msg": "ok", "data": content})
}

func AddClusterInfo(c *gin.Context) {

}

func Map2String(m map[string]string) (result string) {
	list := make([]string, 0)
	for k, v := range m {
		t1 := fmt.Sprintf("%s=%s", k, fmt.Sprint(v))
		list = append(list, t1)
	}
	result = strings.Join(list, ",")
	return
}

type NewEnvInfo struct {
	ClusterId   int    `json:"cluster_id"`
	FromNspName int    `json:"from_nsp_name"`
	Namespace   int    `json:"namespace"`
	SelectedPro string `json:"selected_pro"`
}

func CreateANewEnv(c *gin.Context) {
	newEnvInfo := &NewEnvInfo{}
	err := c.BindJSON(newEnvInfo)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err_code": 20001, "err_msg": "参数绑定错误"})
		return
	}
	devopsdb := database.GetDevopsDb()
	projects, err := devopsdb.GetAllPro(c)
	if err != nil {
		nlog.Error(err)
		c.JSON(200, gin.H{"err_code": 0, "err_msg": err.Error()})
		return
	}

	clusterinfo, err := devopsdb.GetClusterById(c, newEnvInfo.ClusterId)
	if err != nil {
		nlog.Error(err)
		c.JSON(200, gin.H{"err_code": 0, "err_msg": err.Error()})
		return
	}
	opsInfo, errs := addnewtenant.CreateNewTenant(c, addnewtenant.CreateTenantOps{
		FromNameSpace: newEnvInfo.FromNspName,
		ToNameSpace:   newEnvInfo.Namespace,
		Projects:      projects,
		ClusterInfo:   clusterinfo,
	})
	c.JSON(200, gin.H{"err_code": 0, "err_msg": errs, "opsInfo:": opsInfo})
}

type StatusData struct {
	ClusterName    string
	ClusterId      int
	DeploymentName string
	Namespace      string
	Data           string
}

type DeployStatus struct {
	Namespacename  string `form:"nsname"`
	Deploymentname string `form:"dename"`
}

type PodInfo struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	RunTime string `json:"run_time"`
}

func GetDeployStatus(c *gin.Context) {
	var Deploy DeployStatus
	if c.ShouldBindQuery(&Deploy) == nil {
		log.Println(Deploy.Namespacename)
		log.Println(Deploy.Deploymentname)

		var err error
		ctx := context.Background()
		client, err := k8sutil.GetK8sClientById(ctx, 1)
		if err != nil {
			nlog.WithContext(c).Errorf("get k8sclient err:", err)
			c.JSON(http.StatusBadRequest, gin.H{"err_code": 20001, "err_msg": "get k8client" + err.Error()})
			return
		}

		result, err := client.AppsV1().Deployments(Deploy.Namespacename).Get(ctx, Deploy.Deploymentname, metav1.GetOptions{})
		if err != nil {
			nlog.WithContext(c).Error(err)
			c.JSON(http.StatusBadRequest, gin.H{"err_code": 20001, "err_msg": err.Error()})
			return
		}
		resultLabel := result.GetObjectMeta().GetLabels()
		nlog.Info(resultLabel)
		resultString := Map2String(resultLabel)

		podInterface := client.CoreV1().Pods(Deploy.Namespacename)
		podList, err := podInterface.List(ctx, metav1.ListOptions{
			LabelSelector: resultString,
		})
		if err != nil {
			nlog.WithContext(c).Error(err)
			c.JSON(http.StatusBadRequest, gin.H{"err_code": 20001, "err_msg": err.Error()})
			return
		}

		//var podStatus map[string]interface{}
		podInfos := []PodInfo{}
		//用数组的方式，没用单map的方式；主要是考虑有多Pod的情况
		for _, value := range podList.Items {
			_, e := json.Marshal(value)
			if e != nil {
				nlog.WithContext(c).Error(e)
				continue
			}
			podInfo := PodInfo{
				Name:    value.Name,
				Status:  string(value.Status.Phase),
				RunTime: time.Now().Sub(value.Status.StartTime.Time).String(),
			}
			podInfos = append(podInfos, podInfo)
		}

		c.JSON(http.StatusOK, podInfos)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"err_code": 20001, "err_msg": "参数绑定错误"})
		return
	}
}

type CompareInfo struct {
	ConfigId int `json:"config_id"`
}

func CompareConfigInK8sAndApollo(c *gin.Context) {
	compareInfo := CompareInfo{}

	err := c.BindJSON(&compareInfo)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err_code": 0, "err_msg": "参数绑定错误！"})
		return
	}
	res, err := compare.CompareApolloConfigAndK8s(c, compareInfo.ConfigId)
	if err != nil {
		c.JSON(500, gin.H{"err_code": 0, "err_msg": err.Error()})

		return
	}
	c.JSON(http.StatusOK, gin.H{"err_code": 1, "err_msg": "ok", "compare_res": res})

}

type CompareRes struct {
}

func GetApolloConfig(c *gin.Context) {
	configIdStr := c.Query("config_id")
	configId, err := strconv.Atoi(configIdStr)
	if err != nil {
		c.JSON(400, gin.H{"err_code": 0, "err_msg": err.Error()})
		return
	}
	devopsdb := database.GetDevopsDb()
	configMapping, err := devopsdb.GetConfigById(c, configId)
	if err != nil {
		c.JSON(500, gin.H{"err_code": 0, "err_msg": err.Error()})
		return
	}
	nsp, err := devopsdb.GetNamespaceById(c, configMapping.NamespaceId)
	if err != nil {
		c.JSON(500, gin.H{"err_code": 0, "err_msg": err.Error()})
		return
	}
	clusterInfo, err := devopsdb.GetClusterById(c, nsp.K8sClusterId)
	if err != nil {
		c.JSON(500, gin.H{"err_code": 0, "err_msg": err.Error()})
		return
	}
	content, err := apollo.GetYamlFromApollo(c, configMapping.ProjectId, configMapping.ConfigName, clusterInfo.Name, nsp.Name)
	if err != nil {
		c.JSON(500, gin.H{"err_code": 0, "err_msg": err.Error()})
		return
	}
	c.JSON(200, gin.H{"err_code": 1, "err_msg": "ok", "content": content})
}

func SetApolloConfig(c *gin.Context) {
	setConfigInfo := &SetConfigInfo{}
	err := c.BindJSON(setConfigInfo)
	if err != nil {
		c.JSON(400, gin.H{"err_code": 0, "err_msg": err.Error()})
		return
	}
	err = apollo.SetApConfig(c, setConfigInfo.ConfigId, setConfigInfo.Content, setConfigInfo.FileType)
	if err != nil {
		c.JSON(500, gin.H{"err_code": 0, "err_msg": err.Error()})
		return
	}
	c.JSON(200, gin.H{"err_code": 1, "err_msg": "ok"})
}

func PubApolloConfig(c *gin.Context) {
	configIdstr := c.Query("config_id")
	configId, err := strconv.Atoi(configIdstr)
	if err != nil {
		c.JSON(400, gin.H{"err_code": 0, "err_msg": "参数错误！"})
		return
	}
	err = apollo.PubApolloConfig(c, configId)
	if err != nil {
		c.JSON(500, gin.H{"err_code": 0, "err_msg": err.Error()})
		return
	}
	c.JSON(200, gin.H{"err_code": 1, "err_msg": "ok"})
}

type SetConfigInfo struct {
	ConfigId int    `json:"config_id"`
	Content  string `json:"content"`
	FileType string `json:"file_type"`
}

func IsNumber(s string) bool {
	for _, r := range s {
		if 48 <= r && r <= 57 {

		} else {
			return false
		}
	}
	return true
}

func SetContainerSpecialCh(s string) string {
	special := map[rune]bool{
		':': true, '{': true, '}': true, '[': true, ']': true, ',': true, '&': true, '*': true, '#': true, '?': true, '|': true, '-': true, '<': true, '>': true, '=': true, '!': true, '%': true, '"': true, '\'': true,
	}
	for _, r := range s {
		if r == '"' {
			s = "'" + s + "'"
			return s
		}
		if r == '\'' {
			s = `"` + s + `"`
			return s
		}
	}
	for i, r := range s {
		if i == 0 {
			if _, ok := special[r]; ok {
				s = `"` + s + `"`
				return s
			}
		}
	}
	return s
}

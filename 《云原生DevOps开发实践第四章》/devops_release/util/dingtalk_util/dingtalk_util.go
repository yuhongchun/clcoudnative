package dingtalkutil

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"devops_release/config"
	"devops_release/database"
	"github.com/sirupsen/logrus"
)

const CI_MSG = `

`

const MSG = `
[nighting-build]服务变更通知  
**状态**: <font color=${color}>${status}</font>  
**事件类型**: ${event_type}  
**环境**: ${env}  
**关联项目**: ${project_arr}  
**集群信息**: ${cluster}/${namespace}  
**镜像信息**: ${image}  
**关联deployment**: ${deployment_arr}  
**操作的deployment**: ${ops_deployment_arr}  
**操作失败的deployment**: ${failed_deployment_arr}  
**关联的service**: ${servie_arr}  
**apply失败的service**: ${failed_service_arr}  
**关联的config**: ${config_arr}  
**apply失败的config**: ${failed_config_arr}  
**操作人**: ${user} @${user_number}  
**更新时间**: ${time}  
`

const CONFIG_UPDATE_MSG = `
**事件类型**: 配置变更
**关联项目**: ${project_arr}
**关联deployment**: ${deployment_arr}
**重启deployment**: ${restart_deployment_arr}
**未重启deployment**: ${unrestart_deployment_arr}
**重启失败的deployment**: ${restart_failed_deployment_arr}
**操作人**: ${user} @${user_number}
**更新时间**: ${time}
`
const CALLBACK_MSG = `
**事件类型**: ${event_type}
**项目名**: ${project_name}
**操作的deployment**: ${deployment_arr}
**操作失败的deployment**: ${failed_deployment_arr}
**关联的service**: ${servie_arr}
**apply失败的service**: ${failed_service_arr}
**关联的config**: ${config_arr}
**apply失败的config**: ${failed_config_arr}
**操作人**: ${user} @${user_number}
**更新时间**: ${time}
`

type UpdateConfigMsg struct {
	Projects  []string
	EventType []string
}
type CallbackMsg struct {
}

func SendDingtalkBotMsgWithCtx(ctx context.Context) {

}

// func SendDingtalkBotMsgWithOpsInfo(opsinfo *OpsInfo) {

// }

func SetMsg(opsinfo *OpsInfo) string {

	color := "#008000"
	if opsinfo.Status != "success" {
		color = "#FF0000"
	}
	msg := MSG
	msg = strings.ReplaceAll(msg, "${env}", config.ApplicationConfig.Mode)
	msg = strings.ReplaceAll(msg, "${color}", color)
	msg = strings.ReplaceAll(msg, "${image}", opsinfo.Image)
	msg = strings.ReplaceAll(msg, "${status}", opsinfo.Status)
	msg = strings.ReplaceAll(msg, "${project_arr}", fmt.Sprintf("%v", opsinfo.Projects))
	msg = strings.ReplaceAll(msg, "${cluster}", opsinfo.Cluster)
	msg = strings.ReplaceAll(msg, "${namespace}", opsinfo.Namespace)
	msg = strings.ReplaceAll(msg, "${event_type}", opsinfo.EventType)
	msg = strings.ReplaceAll(msg, "${time}", time.Now().Format("2006-01-02 15:04:05 -0700"))
	deploymentArr := ""
	filedDeploymentArr := ""
	for _, da := range opsinfo.Deployments {
		deploymentArr = deploymentArr + da.Name + " "
		if len(da.Err) != 0 {
			filedDeploymentArr = filedDeploymentArr + da.Name + " err:" + da.Err + "\n"
		}
	}
	msg = strings.ReplaceAll(msg, "${deployment_arr}", strings.TrimSpace(deploymentArr))
	msg = strings.ReplaceAll(msg, "${ops_deployment_arr}", strings.TrimSpace(deploymentArr))
	msg = strings.ReplaceAll(msg, "${failed_deployment_arr}", strings.TrimSpace(filedDeploymentArr))

	serviceArr := ""
	filedServiceArr := ""
	for _, s := range opsinfo.ServiceInfo {
		serviceArr = serviceArr + fmt.Sprintf("%d", s.Id) + " "
		if len(s.Err) != 0 {
			filedServiceArr = filedServiceArr + fmt.Sprintf("%d", s.Id) + " err:" + s.Err + "\n"
		}
	}
	msg = strings.ReplaceAll(msg, "${servie_arr}", strings.TrimSpace(serviceArr))
	msg = strings.ReplaceAll(msg, "${failed_service_arr}", strings.TrimSpace(filedServiceArr))

	configArr := ""
	filedConfigArr := ""
	for _, c := range opsinfo.Configs {
		configArr = configArr + c.Name + " "
		if len(c.Err) != 0 {
			filedConfigArr = filedConfigArr + c.Name + " err:" + c.Err + "\n"
		}
	}
	msg = strings.ReplaceAll(msg, "${config_arr}", configArr)
	msg = strings.ReplaceAll(msg, "${failed_config_arr}", filedConfigArr)

	msg = strings.ReplaceAll(msg, "${user}", opsinfo.OpsUser.Name)
	msg = strings.ReplaceAll(msg, "${user_number}", opsinfo.OpsUser.Phone)
	msg = strings.ReplaceAll(msg, "${time}", opsinfo.UpdateTime)
	if len(opsinfo.ErrMsg) != 0 {
		msg = msg + "\n" + "**错误信息**: " + opsinfo.ErrMsg
	}
	return msg
}
func StartOpsInfoSpan(ctx context.Context) (*OpsInfo, context.Context) {
	opsInfo := &OpsInfo{}
	c := context.WithValue(ctx, "opsinfo", opsInfo)
	return opsInfo, c
}

func GetOpsInfoWithContext(ctx context.Context) *OpsInfo {
	opsInfo, ok := ctx.Value("opsinfo").(*OpsInfo)
	if !ok || opsInfo == nil {
		logrus.Error("opsinfo 获取失败！")
		return &OpsInfo{}
	}
	return opsInfo
}

func SendReleaseMsg(ctx context.Context, opsInfo *OpsInfo) {
	msg := SetMsg(opsInfo)
	var dingAt *DingAt
	dingAt = &DingAt{
		AtMobiles: []string{opsInfo.OpsUser.Phone},
	}
	msgTitle := "[" + "nighting-build]" + "服务变更提示"
	dingMsg := DingMsg{
		MsgType: "markdown",
		Markdown: &Markdown{
			Title: msgTitle,
			Text:  msg,
		},
		At: dingAt,
	}
	devopsdb := database.GetDevopsDb()
	for _, p := range opsInfo.Projects {
		dingtalkBotUrlOfPro, err := devopsdb.SelectDingtalkBotByPro(ctx, p.Id)
		if err != nil {
			logrus.Error(err)
			continue
		}
		for _, durl := range dingtalkBotUrlOfPro {

			SendDingtalkBotMsg(durl.DingTalkBotHook, dingMsg)
		}
	}

}

func SendDingtalkBotMsg(dingtalkBotUrl string, msg DingMsg) {

	var (
		result struct {
			ErrCode int64  `json:"errcode"` //nolint:tagliatelle
			ErrMsg  string `json:"errmsg"`  //nolint:tagliatelle
		}
	)

	urlStr := dingtalkBotUrl
	jsonStr, _ := json.Marshal(msg)
	req, err := http.NewRequestWithContext(context.Background(), "POST", urlStr, bytes.NewBuffer(jsonStr))
	if err != nil {
		logrus.Errorf("Error: build dingtalk request failed, err: %s", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logrus.Errorf("Error: send dingtalk msg failed, err: %s", err)
		return
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			logrus.Errorf("Error: close dingtalk response body failed, err: %s", err)
		}
	}()

	body, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal([]byte(string(body)), &result)

	if err != nil {
		logrus.Errorf("Error: unmarshal dingtalk response failed, err: %s", err)
		return
	}
	if result.ErrCode != 0 {
		logrus.Errorf("Error: send dingtalk msg failed, err: %s", result.ErrMsg)
		return
	}
}

type OpsInfo struct {
	Status      string
	Image       string
	Projects    []ProjectInfo
	EventType   string
	Cluster     string
	Namespace   string
	Deployments []DeploymentInfo
	Configs     []ConfigInfo
	ServiceInfo []ServiceInfo
	OpsUser     User
	UpdateTime  string
	ErrMsg      string
}

type ProjectInfo struct {
	Id   int
	Name string
}

type DeploymentInfo struct {
	Id   int
	Name string
	Err  string
}
type ConfigInfo struct {
	Id   int
	Name string
	Err  string
}
type ServiceInfo struct {
	Id   int
	Name string
	Err  string
}
type User struct {
	Name  string
	Phone string
	Email string
}

type DingMsg struct {
	MsgType  string    `json:"msgtype"` //nolint:tagliatelle
	Text     *TextMsg  `json:"text,omitempty"`
	Markdown *Markdown `json:"markdown,omitempty"`
	At       *DingAt   `json:"at,omitempty"`
}

type TextMsg struct {
	Content string `json:"content"`
}

type Markdown struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

type DingAt struct {
	AtMobiles []string `json:"atMobiles"` //nolint:tagliatelle
	IsAtAll   bool     `json:"isAtAll"`   //nolint:tagliatelle
}

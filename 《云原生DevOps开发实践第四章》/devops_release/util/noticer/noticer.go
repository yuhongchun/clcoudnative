package noticer

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"devops_release/config"
	"github.com/sirupsen/logrus"
)

const (
	RestartMsg = `
项目名：%s
操作类型：%s
环境：%s
集群：%s
命名空间：%s
重启信息：%s
`
	ConfigEmptyMsg = `
项目名：%s
操作类型：%s
环境：%s
集群：%s
命名空间：%s
重启信息：没有从配置中心中读取到配置，将跳过configmap apply，请检查配置中心中是否有相应配置和项目名是否正确！
`
)

type msgPost struct {
	IdType  string   `json:"id_type"`
	To      []string `json:"to"`
	Content string   `json:"content"`
	Way     []string `json:"way"`
}
type msgRes struct {
	Message string `json:"message"`
}

func Send(type_ string, to []string, content string, way []string) string {
	msg := msgPost{
		IdType:  type_,
		To:      to,
		Content: content,
		Way:     way,
	}
	bdata, err := json.Marshal(msg)
	if err != nil {
		logrus.Error("消息序列化失败,err:", err)
		return ""
	}
	url := strings.TrimRight(config.NoticerConfig.Host, "/") + "/send"
	res, err := http.Post(url, "application/json", bytes.NewReader(bdata))
	if err != nil {
		logrus.Error("消息post失败,err:", err)
		return ""
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			logrus.Error("钉钉http关闭失败,err:", err)
		}
	}()
	resData, _ := ioutil.ReadAll(res.Body)
	var resJson msgRes
	err = json.Unmarshal(resData, &resJson)
	if err != nil {
		logrus.Error("发送消息结果反序列化失败,err:", err)
		return ""
	}
	return resJson.Message
}
func GetTemplate() string {
	template := `
%s
状态: <font color=%s size=3>%s</font>
项目名: %s
commit_id: %s
docker_tag: %s
commit提交人: %s @%s
项目链接: [gitlab链接](%s)
created_at: %v
finished_at: %v
commit_msg: %s
err_msg: %s
`
	return template
}

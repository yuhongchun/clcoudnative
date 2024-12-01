package gitlabapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/xanzy/go-gitlab"
	"devops_build/config"
	httputil "devops_build/util/http_util"
	"github.com/sirupsen/logrus"
)

type GitlabApi struct {
	PriToken string
}

var GitlabApiObj = GitlabApi{
	PriToken: config.GitlabConfig.PriToken,
}

func (*GitlabApi) GetSingleProject(ctx context.Context, group string, projectName string) (*gitlab.Project, error) {
	encodeUrl := url.PathEscape(group + "/" + projectName)
	url := fmt.Sprintf("http://%s/projects/%s", config.GitlabConfig.ApiHost, encodeUrl)
	header := map[string]string{"PRIVATE-TOKEN": config.GitlabConfig.PriToken}
	body, err := httputil.SendHttpRequest("GET", header, url, nil)
	if err != nil {
		logrus.Error("获取项目信息失败！err:", err)
		return nil, err
	}
	gitlabPro := gitlab.Project{}
	err = json.Unmarshal(body, &gitlabPro)
	if err != nil {
		logrus.Error("json解析错误！")
		return nil, err
	}
	return &gitlabPro, nil
}

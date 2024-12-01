package api

import (
	"context"
	"regexp"
	"strings"

	dbmodel "devops_build/database/model"

	"devops_build/config"
	"devops_build/database"
	"devops_build/internal/model"
	"devops_build/internal/send"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.elastic.co/apm"
)

type MyParams struct {
}

func PipeListening(c *gin.Context) {
	ctx := c.Request.Context()
	span, ctx := apm.StartSpan(ctx, "PipeListening", "POST")
	defer span.End()

	xGitlabEvent := c.Request.Header.Get("X-Gitlab-Event")
	if xGitlabEvent != "Pipeline Hook" {
		logrus.WithContext(ctx).Error("ERROR: XGitlabEvent is not Pipeline Hook")
		return
	}

	params := model.PipelineEvents{}
	if err := c.Bind(&params); err != nil {
		logrus.WithContext(ctx).Errorf("ERROR: Params bind failed, err: %s", err)
		return
	}
	if params.ObjectKind != "pipeline" {
		logrus.WithContext(ctx).Error("ERROR: Object kind not pipeline")
		return
	}

	lables := make(map[string]string)
	lables["project_name"] = params.Project.Name

	if params.ObjectAttributes.Tag {
		lables["project_tag"] = params.ObjectAttributes.Ref
	} else {
		lables["project_branch"] = params.ObjectAttributes.Ref
	}

	status := params.ObjectAttributes.Status
	lables["webhook_status"] = status

	devopsdb := database.GetDevopsDb()
	project, err := devopsdb.GetProjectByGitlabId(ctx, params.Project.ID)
	if err != nil {
		logrus.Error("查询project失败！err:", err)
		return
	}
	//拦截不能发版的分支或者tag
	enabled_flag := false
	if len(config.GitlabConfig.Keywords) == 0 {
		namespace_id := 0
		enabled_flag, namespace_id = RefFilter(ctx, *project, params)
		logrus.Infof("查询到%v的发版路径 namespace:%d", project, namespace_id)
	} else {
		for _, keyword := range config.GitlabConfig.Keywords {
			if strings.Contains(params.ObjectAttributes.Ref, keyword) {
				enabled_flag = true
			}
		}
	}
	if !enabled_flag {
		logrus.Info("refrep不匹配，或者发版路径未启用！")
		return
	}

	if status != "success" {
		if status == "created" || status == "failed" || status == "canceled" || status == "skipped" {
			logrus.Info("status is not success")
		}
		return
	}

	if err != nil {
		logrus.WithContext(ctx).Error("select project err!err:", err)
		return
	}
	if len(project.ProjectToken) == 0 {
		logrus.WithContext(ctx).Errorf("ERROR: Get Web hook token failed, err: %s", err)
		return
	}
	if string(project.ProjectToken) != c.Request.Header.Get("X-Gitlab-Token") {
		logrus.WithContext(ctx).
			Errorf("ERROR: Validate Web hook token failed, request token: %s, should be: %s",
				c.Request.Header.Get("X-Gitlab-Token"), project.ProjectToken)
		return
	}

	err = send.SendRelease(ctx, &params, *project)
	if err != nil {
		logrus.WithContext(ctx).Errorf("Error: send to release failed, err: %s", err)
		params.ObjectAttributes.Status = "error"
		params.IsFailed = true
		params.ErrorMsg = err.Error()
		return
	}

}
// 查找发版路径并且进行检查，如果没有正则匹配则拒绝
func RefFilter(ctx context.Context, project dbmodel.Project, params model.PipelineEvents) (bool, int) {
	devopsdb := database.GetDevopsDb()
	routes, err := devopsdb.GetRoutesRepNamespaceIdByProId(ctx, project.Id)
	if err != nil {
		logrus.Infof("未找到发版路径！project:%v err:%v", project, err)
		return false, 0
	}
	for _, route := range routes {
		logrus.Info(*route)
		r, err := regexp.Compile(route.RefRep)
		if err != nil {
			logrus.Errorf("正则表达式错误！pro:%s env:%s rep：%s", params.Project.Name, route.Channel, route.RefRep)
			continue
		}
		if r.FindString(params.ObjectAttributes.Ref) != "" {
			return true, route.NamespaceId
		}
		if !route.Enabled {
			return false, 0
		}
	}
	return false, 0

}

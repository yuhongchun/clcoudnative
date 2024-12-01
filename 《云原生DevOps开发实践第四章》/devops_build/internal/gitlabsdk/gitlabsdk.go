package gitlabsdk

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"

	"github.com/xanzy/go-gitlab"
	"devops_build/config"
)

func AddWebHook(ctx context.Context, gitlabId int, addHookOps *gitlab.AddProjectHookOptions) error {
	g, err := gitlab.NewClient(config.GitlabConfig.PriToken, gitlab.WithBaseURL("http://gitlab.yiban.io"))
	if err != nil {
		logrus.Error("添加webhook失败,err", err)
		return err
	}
	webhookIsExist := WebHookIsExist(ctx, gitlabId, *addHookOps.URL)
	if webhookIsExist {
		return fmt.Errorf("hook 已经存在！")
	}
	_, _, err = g.Projects.AddProjectHook(gitlabId, addHookOps)
	if err != nil {
		logrus.Error("添加webhook失败！err:", err)
		return err
	}
	return nil
}
func WebHookIsExist(ctx context.Context, gitlabId int, webhookUrl string) bool {
	g, err := gitlab.NewClient(config.GitlabConfig.PriToken, gitlab.WithBaseURL("http://gitlab.example.io"))
	if err != nil {
		logrus.Error(err)
		return false
	}
	project_hooks, _, _ := g.Projects.ListProjectHooks(gitlabId, &gitlab.ListProjectHooksOptions{})
	for _, hook := range project_hooks {
		if hook.URL == webhookUrl {
			return true
		}
	}
	return false
}

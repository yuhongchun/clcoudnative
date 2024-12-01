package gitlab

import (
	"context"
	"devops_build/config"
	gitmodel "devops_build/internal/model"
	"devops_build/internal/release/model"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strings"
	"sync"
)

type GitlabReleaseInfo struct {
}

var ReleaseInfo = &GitlabReleaseInfo{}

func (g *GitlabReleaseInfo) GetDockerInfo(ctx context.Context, params *gitmodel.PipelineEvents) *model.DockerInfo {
	projectId := params.Project.ID
	jobs := params.Builds
	lock := sync.Mutex{}
	wg := sync.WaitGroup{}
	wg.Add(len(jobs))
	var dockerInfo *model.DockerInfo
	for _, job := range jobs {
		jobid := job.ID
		go func() {
			d, err := getDockerUrlFromJobLog(ctx, projectId, jobid)
			if err != nil {
				logrus.WithContext(ctx).Infof("id:%d %v", jobid, err)
			}
			if d != nil {
				lock.Lock()
				dockerInfo = d
				lock.Unlock()
			}
			wg.Done()
		}()
	}
	wg.Wait()
	return dockerInfo
}

/* 获取项目指定作业的日志或跟踪日志
GET /projects/:id/jobs/:job_id/trace
参考GitLab API文档https://docs.gitlab.cn/jh/api/jobs.html
*/
func getDockerUrlFromJobLog(ctx context.Context, projectId int, jobId int) (*model.DockerInfo, error) {
	fmt.Println(jobId)
	url := fmt.Sprintf("http://%s/projects/%d/jobs/%d/trace", config.GitlabConfig.ApiHost, projectId, jobId)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Add("PRIVATE-TOKEN", config.GitlabConfig.PriToken)
	client := http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	logString := string(bodyBytes)
	logStringSplit := strings.Split(logString, "\n")
	for _, line := range logStringSplit {
		if strings.Contains(line, "echo") {
			continue
		}
		// 其逻辑为：如果项目指定作业的日志中包启了enter-devops-dockertag的字段，则走下面的逻辑,否则返回nil空值(即走另外的逻辑，即从deployment表中获取数据)
		if strings.Contains(line, strings.TrimSpace(config.GitlabConfig.CIJobLogDockerKey)) {
			logrus.Info(config.GitlabConfig.CIJobLogDockerKey)
			line = strings.TrimSpace(line)
			lineSplit := strings.Split(line, ":")
			if len(lineSplit) != 3 {
				logrus.WithContext(ctx).Errorf("dockerurl格式错误！")
				return nil, err
			} else {
				// 如果docker仓库是域名开头的，则用下面的逻辑
				//dockerRepo := strings.TrimSpace(lineSplit[1])
				// 如果docker仓库是以IP开头的，则需要用下面的逻辑
				dockerRepo := lineSplit[0]+":"+lineSplit[1]
				dockerTag := strings.TrimSpace(lineSplit[2])
				logrus.Info(dockerRepo,dockerTag)

				return &model.DockerInfo{Repository: dockerRepo, Tag: dockerTag}, nil
			}
		}
	}
	return nil, fmt.Errorf("未在job日志里找到dockertag！")
}

package send

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"devops_build/config"
	nightingrelease "devops_build/controller/nighting_release"
	"devops_build/database"
	dbmodel "devops_build/database/model"
	"devops_build/internal/model"
	release_gitlab "devops_build/internal/release/gitlab"
	"github.com/sirupsen/logrus"
	"go.elastic.co/apm"
	apiappsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/yaml"
)

type Data struct {
	Id          string `json:"id"`
	EventType   string `json:"event_type"`
	ProjectName string `json:"project_name"`
	TagId       string `json:"tag_id"`
	Channel     string `json:"channel"`
	NamespeceId int    `json:"namespace_id"`
	ProjectId   int    `json:"project_id"`
}
type ReleaseRes struct {
	ErrCode int    `json:"err_code"`
	ErrMsg  string `json:"err_msg"`
}

func SendRelease(ctx context.Context, params *model.PipelineEvents, project dbmodel.Project) error {
	span, ctx := apm.StartSpan(ctx, "SendRelease", "func")
	defer span.End()
	devopsdb := database.GetDevopsDb()
	routes, err := devopsdb.GetRoutesRepNamespaceIdByProId(ctx, project.Id)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return err
	}
	data := Data{}
	data.ProjectId = project.Id
	var tagId string
	getNamespaceId := func(tagOrRef string) (int, error) {
		for _, route := range routes {
			fmt.Println(*route)
			r, err := regexp.Compile(route.RefRep)
			if err != nil {
				logrus.Errorf("正则表达式错误！pro:%s env:%s rep：%s", params.Project.Name, route.Channel, route.RefRep)
				continue
			}
			if r.FindString(tagOrRef) != "" {
				return route.NamespaceId, nil
			}
		}
		return 0, fmt.Errorf("tag or ref name is not allowed！")
	}
	if params.ObjectAttributes.Tag {
		tagId = params.ObjectAttributes.Ref
		ref := params.ObjectAttributes.Ref
		curNsp, err := getNamespaceId(strings.ToLower(ref))
		if err != nil {
			return err
		}
		data.TagId = tagId
		data.NamespeceId = curNsp
	} else {
		tagId = params.ObjectAttributes.SHA
		if len(tagId) < 12 {
			return errors.New("tag id illegal")
		}
		tagId = tagId[:12]
		data.TagId = tagId
		curNsp, err := getNamespaceId(params.ObjectAttributes.Ref)
		if err != nil {
			return err
		}
		data.NamespeceId = curNsp
	}
	//如果在build job日志中扫描到了镜像信息，则用此镜像tag
	dockerInfo := release_gitlab.ReleaseInfo.GetDockerInfo(ctx, params)
	if dockerInfo != nil && len(dockerInfo.Repository) != 0 && len(dockerInfo.Tag) != 0 {
		dockertag := fmt.Sprintf("%s:%s", dockerInfo.Repository, dockerInfo.Tag)
		data.TagId = dockertag
		logrus.Info("获取到job_log里的dockertag，使用此tag：", dockertag)
	} else {
		logrus.Info("未获取到job_log里的dockertag，使用默认tag：", data.TagId)
	}
	data.ProjectName = params.Project.Name
	data.EventType = "build"
	imageUrl := ""
	//判断镜像是否有效
	imageIsVaild := false
	if strings.Contains(data.TagId, ":") {
		imageUrl = data.TagId
	} else {

		tempDeployments, err := devopsdb.GetDeploymentsByProIdAndNspIdAndPageSize(ctx, data.ProjectId, data.NamespeceId, 1, 1)
		if err != nil || len(tempDeployments) == 0 {
			return err
		}
		tempDeploymentStruct := &apiappsv1.Deployment{}
		err = yaml.Unmarshal([]byte(tempDeployments[0].Content), tempDeploymentStruct)
		if err != nil {
			return err
		}
		if len(tempDeploymentStruct.Spec.Template.Spec.Containers) == 0 {
			fmt.Println(tempDeployments[0].Content)
			return fmt.Errorf("deployment中的镜像不合法")
		}
		imagepre := strings.Split(tempDeploymentStruct.Spec.Template.Spec.Containers[0].Image, ":")[0]
		imageUrl = imagepre + ":" + data.TagId
	}
	repoName := strings.ReplaceAll(imageUrl, "ccr.ccs.tencentyun.com/", "")
	repoName = strings.Split(repoName, ":")[0]
	repoName = strings.ReplaceAll(repoName, "${__project_name}", data.ProjectName)
	imageList, err := nightingrelease.ImageList(ctx, repoName)
	if err != nil || len(imageList.ImageList) == 0 {
		logrus.Error("获取腾讯云镜像错误或镜像列表为空！")
		return err
	}
	fmt.Println(imageList)
	imageUrlSplit := strings.Split(imageUrl, ":")
	if len(imageUrlSplit) != 2 {
		return fmt.Errorf("镜像url不合法")
	}
	imageTag := strings.TrimSpace(imageUrlSplit[1])
	for _, t := range imageList.ImageList {
		if t == imageTag {
			imageIsVaild = true
		}
	}
	if !imageIsVaild {
		return fmt.Errorf("镜像非法或者不存在！")
	}

	id := tagId + strconv.FormatInt(time.Now().Unix(), 10)
	h := sha1.New()
	h.Write([]byte(id))
	id = hex.EncodeToString(h.Sum(nil))
	data.Id = id

	dataJson, err := json.Marshal(data)
	if err != nil {
		return err
	}
	logrus.WithContext(ctx).Infof("Info: sent to realse data: %s", dataJson)
	fmt.Println("data:", data)
	reqctx, _ := context.WithTimeout(ctx, time.Minute)
	url := config.ApplicationConfig.NightingReleaseUrl
	client := &http.Client{}
	reader := strings.NewReader(string(dataJson))
	req, err := http.NewRequestWithContext(reqctx, "POST", url, reader)
	if err != nil {
		return err
	}
	req.Close = true
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		err := res.Body.Close()
		if err != nil {
			logrus.WithContext(ctx).Errorf("Error: response body close failed, err: %s", err)
		}
	}()
	bodyData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("读取响应错误！err:%s", err)
	}
	releaseRes := &ReleaseRes{}
	err = json.Unmarshal(bodyData, releaseRes)
	if err != nil {
		return fmt.Errorf("json解析错误！err:%s", err)
	}
	if releaseRes.ErrCode == 0 {
		return fmt.Errorf(releaseRes.ErrMsg)
	}
	return nil
}

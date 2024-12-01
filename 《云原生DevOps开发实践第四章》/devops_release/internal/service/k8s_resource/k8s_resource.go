package k8sresource

import (
	"fmt"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tcr "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tcr/v20190924"
	"devops_release/config"
	nlog "github.com/sirupsen/logrus"
)

func GetLastDockerTag(repo string) (string, error) {
	credential := common.NewCredential(
		config.TecentCloudConfig.SecretId, //写在配置文件中
		config.TecentCloudConfig.SecretKey,
	)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "tcr.tencentcloudapi.com"
	client, _ := tcr.NewClient(credential, "ap-guangzhou", cpf)

	request := tcr.NewDescribeImagePersonalRequest()

	request.RepoName = common.StringPtr(repo)
	request.Limit = common.Int64Ptr(30)
	response, err := client.DescribeImagePersonal(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		nlog.Errorf("请求错误 err:%v", err)
		return "", err
	}
	if err != nil {
		panic(err)
	}
	data := response.Response.Data
	tagInfo := data.TagInfo
	if len(tagInfo) == 0 {
		return "", fmt.Errorf("tag is empty")
	}
	lastTagName := tagInfo[0].TagName
	return *lastTagName, nil
}


func GetDockerTagList(repo string) ([]string, error) {
	credential := common.NewCredential(
		config.TecentCloudConfig.SecretId, //写在配置文件中
		config.TecentCloudConfig.SecretKey,
	)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "tcr.tencentcloudapi.com"
	client, _ := tcr.NewClient(credential, "ap-guangzhou", cpf)

	request := tcr.NewDescribeImagePersonalRequest()

	request.RepoName = common.StringPtr(repo)
	request.Limit = common.Int64Ptr(30)
	response, err := client.DescribeImagePersonal(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		nlog.Errorf("请求错误 err:%v", err)
		return nil, err
	}
	if err != nil {
		panic(err)
	}
	data := response.Response.Data
	tagInfos := data.TagInfo
	if len(tagInfos) == 0 {
		return nil, fmt.Errorf("tag list is empty")
	}
	tags := []string{}
	for i, tagInfo := range tagInfos {
		if i > 30 {
			break
		}
		tags = append(tags, *tagInfo.TagName)
	}
	return tags, nil
}

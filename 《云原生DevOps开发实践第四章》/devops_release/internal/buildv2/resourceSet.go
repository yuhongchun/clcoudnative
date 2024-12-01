package buildv2

import (
	"context"
	"fmt"

	"devops_release/database/model"
	"devops_release/database/relational"
	"devops_release/util/apollo"
	logr "github.com/sirupsen/logrus"
)

type ProjectResource struct {
	Configmaps  []string
	Deployments []*model.Deployment
	Services    []*model.Service
	Configs     []*model.Config
}

//获取项目发版资源（deployment,configmap,service）
func (s *ProjectResource) setConfig(ctx context.Context, devopsdb relational.DevopsDb, clusterName string, project model.Project, k8sNamespace model.K8sNamespace, class string) error {
	// env cluster namespace projectname,从apollo获取配置
	logr.Info("依次获取发版资源:configmap/deployment/service")
	logr.Infof("从apollo拉取配置！")
	ch := apollo.GetApChannelFromConfigByClusterName(clusterName)
	if ch == nil {
		logr.Errorf("%s不在apollo配置中！", clusterName)
		return fmt.Errorf("%s不在apollo配置中！", clusterName)
	}
	fmt.Println("获取项目发版资源",[]string{ch.Key, k8sNamespace.Name, project.ProjectName, class})

	//获取deployment与配置的映射关系
	for _, deployment := range s.Deployments {
		// 通过数据库查询deployment相应的config信息
		configs, err := devopsdb.GetConfigByDeploymentId(ctx, deployment.Id)
		if err != nil {
			logr.Errorf("未获取到%s的配置，忽略。。", deployment.DeploymentName)
			continue
		}
		s.Configs = append(s.Configs, configs...)
	}

	//opsMsg := dingtalkutil.GetOpsInfoWithContext(ctx)
	for _, config := range s.Configs {
		logr.Infof("通过Apollo来重新生成应用配置")
		//configmap, err := apollo.GetConfigFromApolloV2(ch.Key, k8sNamespace.Name, *config, class)
		configmap, err := apollo.GetConfigFromApolloV2(ch.Key, k8sNamespace.Name, project.ProjectName, class)
		logr.Info("config信息为:",config)
		if err != nil {
			//opsMsg.Status = "Warning"
			//opsMsg.ErrMsg = "配置" + config.ConfigName + "获取失败！" + err.Error()
			logr.Errorf("获取apollo配置失败！err：%v", err)
		}
		if len(configmap) == 0 {
			//TODO:
			logr.Info("configmap is empty,skip apply info!deploymentId:", project.ProjectName)
		}
		s.Configmaps = append(s.Configmaps, configmap)

	}

	// set configmap

	// set deployment
	if len(s.Deployments) == 0 {
		fmt.Println("projectId:", project.Id)
		logr.WithContext(ctx).Errorf("系统中未找到%s-%s的deployments，或查询出错！", project.ProjectName, k8sNamespace.Name)
		return fmt.Errorf("系统中未找到%s-%s的deployments，或查询出错！", project.ProjectName, k8sNamespace.Name)
	}
	// set service
	services := []*model.Service{}
	for _, deployment := range s.Deployments {
		servicesT, err := devopsdb.GetServicesByDeploymentId(ctx, deployment.Id)
		if err != nil || len(servicesT) == 0 {
			logr.Infof("未找到%s-%s的service，跳过... err:%v", deployment.DeploymentName, k8sNamespace.Name, err)
			continue
		}
		services = append(services, servicesT...)
	}
	s.Services = services
	return nil
}

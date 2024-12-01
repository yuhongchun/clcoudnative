package buildv2

import (
	"context"
	"fmt"
	"strings"

	"devops_release/database"
	"devops_release/database/model"
	"devops_release/database/relational"
	k8sresource "devops_release/internal/service/k8s_resource"
	aesutil "devops_release/util/aes_util"
	dingtalkutil "devops_release/util/dingtalk_util"
	"devops_release/util/noticer"
	"github.com/pkg/errors"
	logr "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type ReleaseInfo struct {
	Id          string `json:"id"`
	EventType   string `json:"event_type"`
	TagId       string `json:"tag_id"`
	Deployments []int  `json:"deployments"`
	ProjectId   int    `json:"project_id"`
	NamespaceId int    `json:"namespace_id"`
}
type UpdateSpec struct {
	K8sClusterName string
	Version        string
	OpsType        string
	ClusterToken   string
	ClusterCa      string
	ClusterUrl     string
	Project        model.Project
	Deployments    []int
	ConfigMap     string
	K8sNamespace   model.K8sNamespace
}

//获取项目资源
func (s *UpdateSpec) GetProjectResource(ctx context.Context, devopsdb relational.DevopsDb) (*ProjectResource, error) {
	if s.Project.ProjectName == "" {
		return nil, errors.New("no project name")
	}
	projectResource := &ProjectResource{}
	deployments, err := devopsdb.GetDeploymentsByProAndNamespace(ctx, s.Project.Id, s.K8sNamespace.Id)
	if err != nil {
		return nil, fmt.Errorf("没有找到deployment!")
	}
	projectResource.Deployments = deployments

	//selected deployment根据上层传进的deployment进行筛选  如果没有则为全选
	if s.Deployments != nil && len(s.Deployments) != 0 {
		selectedDeployment := []*model.Deployment{}
		for _, ds := range projectResource.Deployments {
			for _, dId := range s.Deployments {
				if ds.Id == dId {
					selectedDeployment = append(selectedDeployment, ds)
				}
			}
		}
		projectResource.Deployments = selectedDeployment
	}

	tags := s.Project.ParseTags()
	class := tags["type"]
	err = projectResource.setConfig(ctx, devopsdb, s.K8sClusterName, s.Project, s.K8sNamespace, class)
	if err != nil {
		return nil, err
	}

	return projectResource, nil
}

func (s *UpdateSpec) getRealDeployment(deployment string) string {
	deployment = strings.ReplaceAll(deployment, "${__project_name}", s.Project.ProjectName)
	deployment = strings.ReplaceAll(deployment, "${__k8s_namespace}", s.K8sNamespace.Name)

	return deployment
}
func (s *UpdateSpec) getRealService(service string) string {
	service = strings.ReplaceAll(service, "${__project_name}", s.Project.ProjectName)
	service = strings.ReplaceAll(service, "${__k8s_namespace}", s.K8sNamespace.Name)
	return service
}

func (spec *UpdateSpec) getK8sClient() (*kubernetes.Clientset, error) {
	var k8sInfo K8s_cluster_info
	k8sInfo.Name = spec.K8sClusterName
	k8sInfo.Ca = spec.ClusterCa
	k8sInfo.Token = spec.ClusterToken
	k8sInfo.Url = spec.ClusterUrl
	kubernetes.NewForConfig(&rest.Config{})
	clientset, err := kubernetes.NewForConfig(&rest.Config{
		Host: k8sInfo.Url,
		TLSClientConfig: rest.TLSClientConfig{CAData: []byte(k8sInfo.Ca), ServerName: "kubernetes"},
		BearerToken:     k8sInfo.Token,
	})
	if err != nil {
		return nil, err
	}
	return clientset, nil
}
func (spec *UpdateSpec) updateProject(ctx context.Context) (OpsProResult, error) {
	opsMsg := dingtalkutil.GetOpsInfoWithContext(ctx)
	opsProResult := OpsProResult{
		DeploymentErrs: []string{},
		ServiceErrs:    []string{},
		ConfigMapErrs:  []string{},
	}
	var err error
	if spec == nil {
		opsProResult.Status = "failed"
		opsProResult.Message = "updateProject faild:nil UpdateSpec pointer!"
		return opsProResult, errors.New("updateProject faild:nil UpdateSpec pointer!")
	}
	// connect k8s
	devopsdb := database.GetDevopsDb()
	clientset, err := spec.getK8sClient()
	if err != nil {
		opsProResult.Status = "failed"
		opsProResult.Message = err.Error()
		return opsProResult, err
	}
	resource, err := spec.GetProjectResource(ctx, devopsdb)
	if err != nil {

		opsProResult.Status = "failed"
		opsProResult.Message = err.Error()
		return opsProResult, err
	}
	if len(resource.Services) == 0 && len(resource.Deployments) == 0 {
		opsProResult.Status = "failed"
		opsProResult.Message = err.Error()
		return opsProResult, errors.New("not a valid k8s resource")
	}
	// // replace template
	// resource.Configmap, err = spec.getRealConfig(resource.Configmap, client)
	// if err != nil {
	// 	return err
	// }

	// apply configmap
	resourceConfigmap := resource.Configmaps
	fmt.Printf("最终生成的configMap为%v:",resourceConfigmap)
	if len(resourceConfigmap) == 0 {
		//TODO: 发送消息
		logr.Info("配置为空")
		message := fmt.Sprintf(noticer.ConfigEmptyMsg, spec.Project.ProjectName, "更新服务", spec.K8sNamespace.Name, spec.K8sClusterName, spec.K8sNamespace.Description)
		logr.Info("message:",message)
		//sendToWatchers(ctx, message, spec.Project.Id, spec.K8sNamespace.Name)
	}
	for i, cp := range resourceConfigmap {
		config := resourceConfigmap[i]
		if len(cp) == 0 {
			logr.Info("config is empty ,skip。。。。")
			continue
		}
		opsMsg.Configs = append(opsMsg.Configs, dingtalkutil.ConfigInfo{Id: i})
		err = spec.applyConfigmap(ctx, config, clientset)
		if err != nil {
			opsProResult.Status = "failed"
			opsProResult.Message = err.Error()
			opsProResult.ConfigMapErrs = append(opsProResult.ConfigMapErrs, err.Error())
			opsMsg.Configs[i].Err = err.Error()
			opsMsg.Status = "Warning"
		}
	}

	if len(resource.Deployments) == 0 {
		opsProResult.Message = "" +
			"deployments is empty!"
	}
	opsProResult.Status = "success"
	switch spec.OpsType {
	case "build":
		// apply deployment
		for _, deployment := range resource.Deployments {
			deployMsg := dingtalkutil.DeploymentInfo{
				Id:   deployment.Id,
				Name: deployment.DeploymentName,
			}

			deployment.Content = spec.getRealDeployment(deployment.Content)
			err = spec.applyDeployment(ctx, deployment.Content, clientset)
			if err != nil {
				deployMsg.Err = err.Error()
				opsMsg.Deployments = append(opsMsg.Deployments, deployMsg)
				opsMsg.Status = "Warning"
				opsProResult.DeploymentErrs = append(opsProResult.DeploymentErrs, err.Error())
				logr.WithContext(ctx).Errorf("apply deployment pro:%s deployment:%s  err:%v", spec.Project.ProjectName, deployment.DeploymentName, err)
				continue
			}
			opsMsg.Deployments = append(opsMsg.Deployments, deployMsg)

		}
	case "restart":
		//restart deployment
		for _, deployment := range resource.Deployments {
			deployMsg := dingtalkutil.DeploymentInfo{
				Id:   deployment.Id,
				Name: deployment.DeploymentName,
			}
			d, err := clientset.AppsV1().Deployments(spec.K8sNamespace.Name).Get(ctx, deployment.DeploymentName, metav1.GetOptions{})
			if err == nil && d != nil {
				err = k8sresource.SyncDeploymentToLocal(ctx, deployment.Id, *d)
				if err != nil {
					logr.Errorf("同步deployment失败！")
				}
			}
			deployment.Content = spec.getRealDeployment(deployment.Content)
			err = spec.restartDeployment(ctx, deployment.Content, clientset)
			if err != nil {
				logr.WithContext(ctx).Errorf("restart deployment pro:%s deployment:%s  err:%v", spec.Project.ProjectName, deployment.DeploymentName, err)
				deployMsg.Err = err.Error()
				opsMsg.Status = "Warning"
				opsMsg.Deployments = append(opsMsg.Deployments, deployMsg)
				opsProResult.DeploymentErrs = append(opsProResult.DeploymentErrs, fmt.Sprintf("重启失败%s！err:%v", deployment.DeploymentName, err.Error()))
				continue
			}
			opsMsg.Deployments = append(opsMsg.Deployments, deployMsg)
		}
	case "patch_image":
		for _, deployment := range resource.Deployments {
			deployMsg := dingtalkutil.DeploymentInfo{
				Id:   deployment.Id,
				Name: deployment.DeploymentName,
			}
			d, err := clientset.AppsV1().Deployments(spec.K8sNamespace.Name).Get(ctx, deployment.DeploymentName, metav1.GetOptions{})
			if err != nil {
				logr.Infof("%s not found,please create deployment first.", deployment.DeploymentName)
				opsProResult.DeploymentErrs = append(opsProResult.DeploymentErrs, fmt.Sprintf("没有找到该deployment，请先部署。%s", deployment.DeploymentName))
				continue
			}
			err = spec.patchDeployment(ctx, d, *deployment, clientset)
			if err != nil {
				deployMsg.Err = err.Error()
				opsMsg.Status = "Warning"
				opsMsg.Deployments = append(opsMsg.Deployments, deployMsg)
				opsProResult.DeploymentErrs = append(opsProResult.DeploymentErrs, err.Error())
				logr.WithContext(ctx).Errorf("patch deployment pro:%s deployment:%s err:%v", spec.Project.ProjectName, deployment.DeploymentName, err)
				continue
			}
			opsMsg.Deployments = append(opsMsg.Deployments, deployMsg)
		}
	case "patch_deployment":
		// apply deployment
		for _, deployment := range resource.Deployments {
			deployMsg := dingtalkutil.DeploymentInfo{
				Id:   deployment.Id,
				Name: deployment.DeploymentName,
			}

			deployment.Content = spec.getRealDeployment(deployment.Content)
			err = spec.applyDeployment(ctx, deployment.Content, clientset)
			if err != nil {
				deployMsg.Err = err.Error()
				opsMsg.Deployments = append(opsMsg.Deployments, deployMsg)
				opsMsg.Status = "Warning"
				opsProResult.DeploymentErrs = append(opsProResult.DeploymentErrs, err.Error())
				logr.WithContext(ctx).Errorf("apply deployment pro:%s deployment:%s  err:%v", spec.Project.ProjectName, deployment.DeploymentName, err)
				continue
			}
			opsMsg.Deployments = append(opsMsg.Deployments, deployMsg)
		}
	default:
		return opsProResult, fmt.Errorf("not implements")
	}

	// apply service
	for _, service := range resource.Services {
		serviceMsg := dingtalkutil.ServiceInfo{
			Id: service.Id,
		}
		service.Content = spec.getRealService(service.Content)
		err = spec.applyService(ctx, service.Content, clientset)
		if err != nil {
			opsMsg.Status = "Warning"
			serviceMsg.Err = err.Error()
			opsProResult.ServiceErrs = append(opsProResult.ServiceErrs, err.Error())
			logr.WithContext(ctx).Errorf("apply service pro:%s  service:%s err:%v", spec.Project.ProjectName, service.Name, err)
		}
		opsMsg.ServiceInfo = append(opsMsg.ServiceInfo, serviceMsg)

	}
	return opsProResult, nil
}

// restartProject: restart deployment if config file has changed.
// stage1: get config.
// stage2: get apply configmap
// stage3: restart deployment
// stage4: update svc
func (s *UpdateSpec) restart(ctx context.Context) error {
	fmt.Printf("重启的Deployments名字为: %v",s.Deployments)
	opsMsg := dingtalkutil.GetOpsInfoWithContext(ctx)
	var err error
	if s == nil {
		return errors.New("restartProject faild: nil restartProject pointer!")
	}

	devopsdb := database.GetDevopsDb()
	// connect k8s
	clientset, err := s.getK8sClient()
	if err != nil {
		return err
	}
	resource, err := s.GetProjectResource(ctx, devopsdb)
	if err != nil {
		return err
	}
	if len(resource.Configmaps) == 0 && len(resource.Services) == 0 && len(resource.Deployments) == 0 {
		return errors.New("not a valid k8s resource")
	}

	// // replace template
	// resource.Configmap, err = s.getRealConfig(resource.Configmap, client)
	// if err != nil {
	// 	return err
	// }

	// apply configmap
	if len(resource.Configmaps) == 0 {
		logr.Info("配置为空")
		message := fmt.Sprintf(noticer.ConfigEmptyMsg, s.Project.ProjectName, "重启服务", s.K8sNamespace.Name, s.K8sClusterName, s.K8sNamespace.Description)
		logr.Info("message:",message)
		//sendToWatchers(ctx, message, s.Project.Id, s.K8sNamespace.Name)
	}

	//找出configmap关联的deployment并重启
	for i, cfg := range resource.Configs {
		if !cfg.RestartAfterPub {
			continue
		}
		configMsg := dingtalkutil.ConfigInfo{
			Id:   cfg.Id,
			Name: cfg.ConfigName,
		}


		err = s.applyConfigmap(ctx, resource.Configmaps[i], clientset)
		if err != nil {
			opsMsg.Status = "Warning"
			configMsg.Err = err.Error()
			opsMsg.Configs = append(opsMsg.Configs, configMsg)
			continue
		}

		opsMsg.Configs = append(opsMsg.Configs, configMsg)
		deployMsg := dingtalkutil.DeploymentInfo{
			Id: cfg.DeploymentId,
		}
		deployment, err := devopsdb.GetDeploymentById(ctx, cfg.DeploymentId)
		if err != nil {

			deployMsg.Err = "未获取到deployment"
			logr.Error("未获取到deployment")
			opsMsg.Deployments = append(opsMsg.Deployments, deployMsg)
			continue
		}

		deployMsg.Id = deployment.Id
		deployMsg.Name = deployment.DeploymentName
		logr.Infof("deployMsgId %v",deployMsg.Id)
		//同步并重启deployment
		d, err := clientset.AppsV1().Deployments(s.K8sNamespace.Name).Get(ctx, deployment.DeploymentName, metav1.GetOptions{})
		if err == nil && d != nil && len(d.Name) != 0 {
			err = k8sresource.SyncDeploymentToLocal(ctx, deployment.Id, *d)
			if err != nil {
				logr.Errorf("同步deployment失败！")
			}
		} else {
			logr.Infof("同步deployment失败！K8S端deployment未部署")
			deployMsg.Err = "同步deployment失败！K8S端deployment未部署"
			opsMsg.Status = "Warnging"
			opsMsg.Deployments = append(opsMsg.Deployments, deployMsg)
			continue
		}

		//变量替换
		deployment.Content = s.getRealDeployment(deployment.Content)
		err = s.restartDeployment(ctx, deployment.Content, clientset)
		if err != nil {
			opsMsg.Status = "Warning"
			deployMsg.Err = err.Error()
			logr.WithContext(ctx).Errorf("restart deployment pro:%s deployment:%s  err:%v", s.Project.ProjectName, deployment.DeploymentName, err)
			opsMsg.Deployments = append(opsMsg.Deployments, deployMsg)
			continue
		}

	}
	// apply service
	for _, service := range resource.Services {
		serviceMsg := dingtalkutil.ServiceInfo{
			Id: service.Id,
		}
		err = s.applyService(ctx, service.Content, clientset)
		if err != nil {
			serviceMsg.Err = err.Error()
			opsMsg.Status = "Warning"
			opsMsg.ServiceInfo = append(opsMsg.ServiceInfo, serviceMsg)
			logr.WithContext(ctx).Errorf("apply service pro:%s service:%s  err:%v", s.Project.ProjectName, service.Name, err)
		}
		opsMsg.ServiceInfo = append(opsMsg.ServiceInfo, serviceMsg)
	}
	return nil
}

func BuildProject(ctx context.Context, releaseInfo ReleaseInfo) (OpsInfo, error) {
	opsMsg, ctx := dingtalkutil.StartOpsInfoSpan(ctx)
	opsMsg.Status = "success"
	opsMsg.EventType = releaseInfo.EventType
	opsMsg.Image = releaseInfo.TagId
	devopsdb := database.GetDevopsDb()
	project, err := devopsdb.GetProjectById(ctx, releaseInfo.ProjectId)
	if err != nil {
		logr.Errorf("未在系统中找到project id:%d！联系devops人员配置。err：%v", releaseInfo.ProjectId, err)
		return OpsInfo{Status: "failed"}, fmt.Errorf("未在系统中找到project id:%d！联系devops人员配置。err：%v", releaseInfo.ProjectId, err)
	}
	opsMsg.Projects = append(opsMsg.Projects, dingtalkutil.ProjectInfo{Id: project.Id, Name: project.ProjectName})
	k8sNamespace, err := devopsdb.GetNamespaceById(ctx, releaseInfo.NamespaceId)
	if err != nil {
		logr.Errorf("未在系统中找到namespace id:%d！联系devops人员配置。err：%v", releaseInfo.NamespaceId, err)
		opsMsg.Status = "error"
		opsMsg.ErrMsg = err.Error()
		//dingtalkutil.SendReleaseMsg(ctx, opsMsg)
		return OpsInfo{Status: "failed"}, errors.Errorf("未在系统中找到namespace id:%d！联系devops人员配置。err：%v", releaseInfo.NamespaceId, err)
	}

	projectName := project.ProjectName
	deployments := releaseInfo.Deployments
	eventType := releaseInfo.EventType
	tag_id := releaseInfo.TagId
	opsInfo := OpsInfo{
		ProjectName:   project.ProjectName,
		Channel:       k8sNamespace.Name,
		OpsProResults: []OpsProResult{},
		Type:          eventType,
	}
	var updatespec UpdateSpec
	updatespec.Version = tag_id
	updatespec.Project = *project
	updatespec.Deployments = deployments
	updatespec.OpsType = eventType
	updatespec.Project = *project
	updatespec.K8sNamespace = *k8sNamespace
	routes := getReleaseRoutes(ctx, project.Id, releaseInfo.NamespaceId)
	if len(routes) == 0 {
		opsInfo.Status = "failed"
		return opsInfo, fmt.Errorf("系统中没有%s-%s的发版路径！请添加！", projectName, k8sNamespace.Name)
	}
	logr.Infof("重新部署%s:%s", k8sNamespace.Name, projectName)
	for _, route := range routes {
		cluster, err := devopsdb.GetClusterById(ctx, route.ClusterId)
		if err != nil {
			logr.Errorf("未找到pro:%s nsp:%s,cluster：%d  err:%v", projectName, route.NspName, route.ClusterId, err)
			continue
		}
		updatespec.K8sClusterName = cluster.Name
		updatespec.ClusterCa = cluster.Ca
		updatespec.ClusterToken = cluster.Token
		if len(cluster.Ca) == 0 || len(cluster.Token) == 0 {
			realCa, err := aesutil.DecryptString(cluster.EncryptCa)
			realToken, err := aesutil.DecryptString(cluster.EncryptToken)
			if err != nil {
				logr.Error("解密集群字段错误！")
			}
			updatespec.ClusterCa = realCa
			updatespec.ClusterToken = realToken
		}
		updatespec.ClusterUrl = cluster.Url
		opsresult, err := updatespec.updateProject(ctx)
		if err != nil {
			logr.WithContext(ctx).Error(err)
		}
		opsMsg.Cluster = cluster.Name
		opsMsg.Namespace = route.NspName
		opsresult.ClusterName = cluster.Name
		opsresult.Namespace = route.NspName
		opsInfo.OpsProResults = append(opsInfo.OpsProResults, opsresult)
	}
	opsInfo.Status = "success"
	dingtalkutil.SendReleaseMsg(ctx, opsMsg)
	return opsInfo, nil
}

// RestartProject
/*
func RestartProject(ctx context.Context, projectName string, tag_id string, channel string) error {
	devopsdb := database.GetDevopsDb()
	channels := strings.Split(channel, ",")
	if len(channel) == 0 {
		return errors.New("error channel format")
	}
	project, err := devopsdb.GetProjectByName(ctx, projectName)
	if err != nil || project.Id == 0 {
		nlog.Errorf("未在系统中找到%s！联系devops人员配置。err：%v", projectName, err)
		return nil
	}
	var updatespec UpdateSpec
	updatespec.Version = tag_id
	updatespec.Project = *project

	routes := getReleaseRoutes(ctx, project.Id, channels)
	if len(routes) == 0 {
		return fmt.Errorf("系统中没有%s-%s的发版路径！请添加！", projectName, channel)
	}
	nlog.Infof("重新部署%s:%s", channel, projectName)
	for _, route := range routes {
		cluster, err := devopsdb.GetClusterById(ctx, route.ClusterId)
		if err != nil {
			nlog.Errorf("未找到pro:%s-%s,cluster：%s", projectName, route.Channel, route.ClusterId)
			continue
		}
		updatespec.Channel = route.Channel
		updatespec.K8sClusterName = cluster.Name
		updatespec.Namespace = route.NspName
		updatespec.ClusterCa = cluster.Ca
		updatespec.ClusterToken = cluster.Token
		updatespec.ClusterUrl = cluster.Url
		err = updatespec.restart(ctx)
		if err != nil {
			nlog.WithContext(ctx).Error(err)
		}
	}

	// buildmap := getPlan(projectName)
	// nlog.Infof("重新部署%s:%s", channel, projectName)
	// for _, c := range channels {
	// 	fmt.Println(buildmap)
	// 	updatespec.Channel = c
	// 	v, ok := buildmap[c]
	// 	if !ok {
	// 		return fmt.Errorf("not found in build map!env:%s", c)
	// 	}
	// 	for _, k8s_cluster := range v {
	// 		updatespec.K8sClusterName = k8s_cluster.Name
	// 		for _, ns := range k8s_cluster.Namespace {
	// 			updatespec.Namespace = ns
	// 			err := updatespec.restart(ctx)
	// 			if err != nil {
	// 				nlog.WithContext(ctx).Error(err)
	// 			}
	// 		}
	// 	}
	// }
	return nil
}
*/

func PatchDeployment() {
}

//读取apollo配置并重启服务
func RestartWithConfig(ctx context.Context, k8sCluster string, nsp_name string, configName string) error {
	opsMsg, ctx := dingtalkutil.StartOpsInfoSpan(ctx)
	opsMsg.EventType = "配置变更"
	opsMsg.Cluster = k8sCluster
	opsMsg.Namespace = nsp_name
	opsMsg.Status = "success"
	logr.Info("配置变更，重启服务！")
	devopsdb := database.GetDevopsDb()
	clusterId, err := devopsdb.GetClusterIdByName(ctx, k8sCluster)
	if err != nil {
		logr.Error("获取集群信息失败！")
		opsMsg.Status = "error"
		opsMsg.ErrMsg = "获取集群信息失败！"
		dingtalkutil.SendReleaseMsg(ctx, opsMsg)
		return err
	}
	nsp, err := devopsdb.GetNspByClusterAndName(ctx, clusterId.Id, nsp_name)
	if err != nil {
		logr.Error("获取namespace信息失败！")
		opsMsg.Status = "error"
		opsMsg.ErrMsg = "获取命名空间信息失败！"
		//dingtalkutil.SendReleaseMsg(ctx, opsMsg)
		return err
	}

	logr.Info("打印集群及命名空间等信息")
	//fmt.Println(clusterId.Id,nsp_name)
	//fmt.Println(nsp.Id,configName)

	configMappings, err := devopsdb.GetConfigByNamespaceIdAndConfigName(ctx, nsp.Id, configName)
	if err != nil {
		opsMsg.Status = "error"
		opsMsg.ErrMsg = fmt.Sprintf("获取配置[%s]失败！", configName)
		return err
	}

	//for _, configMapping := range configMappings {
	//	fmt.Println("configMap集合为:",configMapping)
	//}

	clusterInfo, err := devopsdb.GetClusterById(ctx, clusterId.Id)
	if err != nil {
		return err
	}

	k8sNamespace, err := devopsdb.GetNspByClusterAndName(ctx, clusterId.Id, nsp_name)
	if err != nil {
		return err
	}

	var updatespec UpdateSpec
	//设置关联的deployment和集群信息
	updatespec.K8sClusterName = clusterInfo.Name
	updatespec.ClusterCa = clusterInfo.Ca
	updatespec.ClusterToken = clusterInfo.Token
	if len(clusterInfo.Ca) == 0 || len(clusterInfo.Token) == 0 {
		//解密
		realCa, err := aesutil.DecryptString(clusterInfo.EncryptCa)
		realToken, err := aesutil.DecryptString(clusterInfo.EncryptToken)
		if err != nil {
			logr.Error("解密token错误！")
		}
		updatespec.ClusterCa = realCa
		updatespec.ClusterToken = realToken
	}
	updatespec.ClusterUrl = clusterInfo.Url
	updatespec.K8sNamespace = *k8sNamespace

	//重启应用配置的逻辑
	for _, configMapping := range configMappings {
		//if !configMapping.RestartAfterPub {
		//	continue
		//}
		configMsg := dingtalkutil.ConfigInfo{
			Id:   configMapping.Id,
			Name: configMapping.ConfigName,
		}
		project, err := devopsdb.GetProjectById(ctx, configMapping.ProjectId)
		if err != nil {
			logr.Error("找不到项目！projectId:", project.Id)

			configMsg.Err = "项目不存在！"
			opsMsg.Configs = append(opsMsg.Configs, configMsg)
			continue
		}
		opsMsg.Projects = append(opsMsg.Projects, dingtalkutil.ProjectInfo{
			Id:   project.Id,
			Name: project.ProjectName,
		})
		updatespec.Project = *project
		updatespec.Deployments = []int{configMapping.DeploymentId}
		err = updatespec.restart(ctx)
		if err != nil {
			logr.Errorf("重启服务失败！pro:%s ")
			//TODO: 发送通知
			message := fmt.Sprintf(noticer.RestartMsg, configName, "配置变更", nsp_name, k8sCluster, nsp_name, err.Error())
			logr.Info(message)
			opsMsg.Status = "error"
			opsMsg.ErrMsg = "重启服务失败！"
			continue
		} else {
			logr.Info("重启服务成功！")
			message := fmt.Sprintf(noticer.RestartMsg, configName, "配置变更", nsp_name, k8sCluster, nsp_name, "正常")
			logr.Info(message)
		}
	}
	fmt.Printf("opsInfo:%#v\n", opsMsg)
	//dingtalkutil.SendReleaseMsg(ctx, opsMsg)
	return nil
}

// 这段逻辑暂时也没用到，可以去掉
func sendToWatchers(ctx context.Context, message string, projectId int, nspName string) {
	devopsdb := database.GetDevopsDb()
	watchers, err := devopsdb.GetWatcherUUIDByNspNameAndProId(ctx, projectId, nspName)
	watchersUuids := []string{}
	if err != nil || watchers == nil {
		logr.Error("查询watchers错误！err:", err)
	} else {
		for _, watcher := range watchers {
			watchersUuids = append(watchersUuids, watcher.UserUUId)
		}
	}
	res := noticer.Send("uuid", watchersUuids, message, []string{})
	logr.Info(res)
}

package addnewtenant

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"devops_release/config"
	"devops_release/database"
	"devops_release/database/model"
	"devops_release/internal/buildv2"
	k8sresource "devops_release/internal/service/k8s_resource"
	"devops_release/util/apollo"
	nlog "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/yaml"
	appsv1 "k8s.io/client-go/applyconfigurations/apps/v1"
)

type CreateTenantOps struct {
	ClusterInfo       *model.K8sClusterInfo
	FromChannel       string
	ToChannel         string
	FromNameSpace     int
	ToNameSpace       int
	FromNameSpaceName string
	ToNamespaceName   string
	Projects          []*model.Project
	ApolloAppId       string
}

func CopyAllConfig(ctx context.Context, ops CreateTenantOps) error {
	devopsdb := database.GetDevopsDb()
	fromNamespaceK8s, err := devopsdb.GetNamespaceById(ctx, ops.FromNameSpace)
	toNamespaceK8s, err := devopsdb.GetNamespaceById(ctx, ops.ToNameSpace)
	if err != nil {
		return fmt.Errorf("查询naemspace出错！err:%v", err)
	}
	ops.FromNameSpaceName = fromNamespaceK8s.Name
	ops.ToNamespaceName = toNamespaceK8s.Name
	apolloOpenapiFrom := apollo.OpenApi{}
	apolloOpenapiTo := apollo.OpenApi{}
	apolloOpenapiTo.ClusterName = toNamespaceK8s.Name
	apolloOpenapiFrom.ClusterName = fromNamespaceK8s.Name

	for _, apChannel := range config.ApolloConfig.Channel {
		if apChannel.K8sCluster == ops.ClusterInfo.Name {
			keySplit := strings.Split(apChannel.Key, "->")
			if len(keySplit) != 2 {
				return fmt.Errorf("apollo 配置错误，导入失败！")
			}
			env := keySplit[1]

			apolloOpenapiFrom.AddressOpenapi = apChannel.AddressOpenapi

			apolloOpenapiTo.AddressOpenapi = apChannel.AddressOpenapi

			err := apolloOpenapiTo.AddNameSpace(ctx, toNamespaceK8s.Name)
			if err != nil {
				nlog.Error("创建命名空间失败！err:", err)
			}
			for _, app := range apChannel.Apps {
				if app.Type == "tenant" {
					apolloOpenapiFrom.Appid = app.Id
					apolloOpenapiFrom.Token = app.Token
					apolloOpenapiFrom.Env = env

					apolloOpenapiTo.Appid = app.Id
					apolloOpenapiTo.Token = app.Token
					apolloOpenapiTo.Env = env
					for _, project := range ops.Projects {
						apolloOpenapiFrom.NamespaceName = project.ProjectName
						apolloOpenapiTo.NamespaceName = project.ProjectName
						err := apolloOpenapiTo.AddNameSpace(ctx, project.ProjectName)
						if err != nil {
							nlog.Error("创建namespace失败！err：", err)
						}
						nspInfo, err := apolloOpenapiFrom.GetNamespaceInfo(project.ProjectName)
						if err != nil {
							fmt.Println(ops.FromNameSpace)
							fmt.Println(ops.ToNameSpace)
							nlog.Error("获取namespace失败！", err)
							continue
						}
						items := nspInfo.Items
						errs := apolloOpenapiTo.AddItems(items)
						if len(errs) != 0 {
							nlog.Info("添加部分items失败！err:", errs)
						}
						apolloOpenapiTo.NamespaceName = project.ProjectName
						err = apolloOpenapiTo.ReleaseConfig()
						if err != nil {
							nlog.Error("release config error!err:", err)
						}
					}
				}
			}
		}

	}
	return nil

}

func StartProjects(ctx context.Context, ops CreateTenantOps) ([]buildv2.OpsInfo, []error) {
	errs := []error{}
	devopsdb := database.GetDevopsDb()
	projects := ops.Projects
	var opsInfos []buildv2.OpsInfo
	for _, project := range projects {
		deployment, err := devopsdb.GetDeploymentsByProAndNamespace(ctx, project.Id, ops.ToNameSpace)
		if err != nil || len(deployment) == 0 {
			nlog.Errorf("start %s faild!err:%v", project.ProjectName, err)
			errs = append(errs, err)
		}
		deploymentYaml := appsv1.DeploymentApplyConfiguration{}
		d := yaml.NewYAMLToJSONDecoder(bytes.NewBufferString(deployment[0].Content))
		e := d.Decode(&deploymentYaml)
		if e != nil {
			nlog.Error("deployment 格式错误！err:%s", err)
		}
		image := deploymentYaml.Spec.Template.Spec.Containers[0].Image
		repoName := strings.ReplaceAll(*image, "ccr.ccs.tencentyun.com/", "")
		repoName = strings.Split(repoName, ":")[0]
		lastTag, err := k8sresource.GetLastDockerTag(repoName)
		if err != nil {
			nlog.Error("获取最新的dockertag错误！", err)
		}
		opsInfo, err := buildv2.BuildProject(ctx, buildv2.ReleaseInfo{
			ProjectId:   project.Id,
			TagId:       lastTag,
			NamespaceId: ops.ToNameSpace,
			Deployments: nil,
			EventType:   "build",
		})
		opsInfos = append(opsInfos, opsInfo)
		if err != nil {
			nlog.Errorf("start %s faild!err:%v", project.ProjectName, err)
			errs = append(errs, err)
		}
	}
	return opsInfos, errs
}
func CopyDeploymentsAndServices(ctx context.Context, ops CreateTenantOps) []error {
	devopsdb := database.GetDevopsDb()
	errs := []error{}
	for _, project := range ops.Projects {
		deployments, err := devopsdb.GetDeploymentsByProAndNamespace(ctx, project.Id, ops.FromNameSpace)
		if err != nil {
			nlog.Errorf("copy %s deployment err from %s！err:%v", project.ProjectName, ops.FromChannel, err)
			errs = append(errs, err)
			continue
		}
		for _, deployment := range deployments {
			fromServices, err := devopsdb.GetServicesByDeploymentId(ctx, deployment.Id)
			if err != nil {
				nlog.Errorf("copy %s service from %s error!err:%v", project.ProjectName, ops.FromChannel, err)
			}
			deployment.Id = 0
			deployment.NamespaceId = ops.ToNameSpace
			d, err := devopsdb.InsertIntoDeployment(ctx, *deployment)
			if err != nil {
				nlog.Errorf("copy %s deployment err to %s！err:%v", project.ProjectName, ops.ToChannel, err)
				errs = append(errs, err)
				break
			}
			for _, fromService := range fromServices {
				toService := fromService
				toService.Id = 0
				toService.DeploymentId = d.Id

				_, err := devopsdb.InsertIntoService(ctx, *toService)
				if err != nil {
					nlog.Errorf("copy service error")
				}
			}
		}
	}
	return errs
}

func CopyRoutes(ctx context.Context, ops CreateTenantOps) []error {
	errs := []error{}
	devopsdb := database.GetDevopsDb()
	for _, project := range ops.Projects {
		routes, err := devopsdb.GetRoutesByProIdAndNspId(ctx, project.Id, ops.FromNameSpace)
		if err != nil {
			errs = append(errs, err)
			nlog.Errorf("copy %s routes failed!err:%v", ops.FromChannel, err)
			continue
		}
		for _, route := range routes {
			toRoute := route
			toRoute.Id = 0
			toRoute.RefRep = ""
			toRoute.NspName = ops.ToNamespaceName
			toRoute.NamespaceId = ops.ToNameSpace
			_, err := devopsdb.InsertIntoRoutes(ctx, *route)
			if err != nil {
				nlog.Errorf(err.Error())
			}
		}
	}
	return errs
}

func CreateNewTenant(ctx context.Context, ops CreateTenantOps) ([]buildv2.OpsInfo, []error) {
	var opsInfos []buildv2.OpsInfo
	err := CopyAllConfig(ctx, ops)
	if err != nil {
		nlog.Info("copy all config failed!err:", err)
	}
	errs := CopyDeploymentsAndServices(ctx, ops)
	if len(errs) != 0 {
		return opsInfos, errs
	}
	errs = CopyRoutes(ctx, ops)
	if len(errs) != 0 {
		return opsInfos, errs
	}
	opsInfos, errs = StartProjects(ctx, ops)
	if len(errs) != 0 {
		return opsInfos, errs
	}
	return opsInfos, []error{}
}
func getK8sClientTemp() {

}

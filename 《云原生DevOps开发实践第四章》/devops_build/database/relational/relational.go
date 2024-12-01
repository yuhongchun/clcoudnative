package relational

import (
	"context"

	. "devops_build/database/model"
)

type DevopsDb interface {
	GetDeploymentsByPageSize(ctx context.Context, page int, size int) ([]*Deployment, error)
	GetServicesByPageSize(ctx context.Context, page int, size int) ([]*Service, error)
	GetProjectsByPageSize(ctx context.Context, page int, size int) ([]*Project, error)
	GetProjectById(ctx context.Context, projectId int) (*Project, error)
	GetRoutesByPageSize(ctx context.Context, page int, size int) ([]*Route, error)

	GetDeploymentsByProAndChannel(ctx context.Context, projectId int, channel string) ([]*Deployment, error)
	GetDeploymentsByProIdAndNspIdAndPageSize(ctx context.Context, proid int, nspid int, page int, size int) ([]*Deployment, error)
	GetDeploymentsCountByProIdAndNspId(ctx context.Context, proid int, nspid int) (int, error)

	GetRoutesByProIdAndNspIdAndPageSize(ctx context.Context, proid int, nspid int, page int, size int) ([]*Route, error)
	GetRoutesCountByProIdAndNspId(ctx context.Context, proid int, nspid int) (int, error)

	GetServicesByDeploymentId(ctx context.Context, deployid int) ([]*Service, error)
	GetServicesCountByDeploymentId(ctx context.Context, deployid int) (int, error)

	GetRoutesByProAndChannel(ctx context.Context, projectId int, channel string) ([]*Route, error)
	GetClusterById(ctx context.Context, id int) (*K8sClusterInfo, error)
	GetNspByClusterAndName(ctx context.Context, clusterId int, name string) (*K8sNamespace, error)
	GetProIdByName(ctx context.Context, name string) (*Project, error)
	GetAllPro(ctx context.Context) ([]*Project, error)
	GetClusterIdByName(ctx context.Context, name string) (*K8sClusterInfo, error)
	GetRoutesByProId(ctx context.Context, projectId int) ([]*Route, error)
	GetProjectByName(ctx context.Context, projectName string) (*Project, error)
	GetProjectsByFuzzyFind(ctx context.Context, fuzzystr string) ([]*Project, error)
	GetEnabledBranchsAndTokenByProName(ctx context.Context, projectName string) (*Project, error)
	GetRoutesRepNamespaceIdByProId(ctx context.Context, projectId int) ([]*Route, error)
	GetAllDeployments(ctx context.Context) ([]*Deployment, error)
	GetAllService(ctx context.Context) ([]*Service, error)
	GetDeploymentByName(ctx context.Context, name string) ([]*Deployment, error)
	GetProjectByGitlabId(ctx context.Context, gitlabId int) (*Project, error)
	GetAllRoutes(ctx context.Context) ([]*Route, error)
	GetDeploymentById(ctx context.Context, deploymentId int) (*Deployment, error)

	InsertIntoRequestHistory(ctx context.Context, requestHistory RequestHistory) (*RequestHistory, error)
	InsertIntoDockerInfo(ctx context.Context, dockerInfo DockerInfo) (*DockerInfo, error)
	InsertIntoDeployment(ctx context.Context, deployment Deployment) (*Deployment, error)
	InsertIntoK8sCluster(ctx context.Context, k8sClusterInfo K8sClusterInfo) (*K8sClusterInfo, error)
	InsertIntoRoutes(ctx context.Context, route Route) (*Route, error)
	InsertIntoK8sNamespace(ctx context.Context, nsp K8sNamespace) (*K8sNamespace, error)
	InsertIntoProjects(ctx context.Context, project Project) (*Project, error)
	InsertIntoService(ctx context.Context, service Service) (*Service, error)
	GetServicesByDeploymentIdAndChannel(ctx context.Context, deploymentId int, channel string) (*Service, error)
	InsertIntoOpsHistory(ctx context.Context, opsHistory OpsHistory) (*OpsHistory, error)

	GetOpsHistoryByPageSize(ctx context.Context, count int, size int) ([]*OpsHistory, error)

	IfExistDeployment(ctx context.Context, deployment Deployment) (bool, error)
	IfExistProject(ctx context.Context, project Project) (bool, error)
	IfExistService(ctx context.Context, service Service) (bool, error)
	IfExistRoute(ctx context.Context, route Route) (bool, error)
	UpdateDeployment(ctx context.Context, deployment Deployment) (bool, error)
	UpdateRoute(ctx context.Context, route Route) (bool, error)
	UpdateProject(ctx context.Context, project Project) (bool, error)
	UpdateService(ctx context.Context, service Service) (bool, error)
	UpdateRouteById(ctx context.Context, route Route) (int, error)

	GetProjectCount(ctx context.Context) (int, error)
	GetDeploymentCount(ctx context.Context) (int, error)
	GetRouteCount(ctx context.Context) (int, error)
	GetServiceCount(ctx context.Context) (int, error)

	GetUserByName(ctx context.Context, name string) (*User, error)
	GetUserByAccount(ctx context.Context, account string) (*User, error)
	InsertIntoUser(ctx context.Context, user User) (*User, error)

	GetCluster(ctx context.Context) ([]*K8sClusterInfo, error)
	GetNspByClusterId(ctx context.Context, clusterid int) ([]*K8sNamespace, error)
	UpdateK8sCluster(ctx context.Context, k8sCluster K8sClusterInfo) (bool, error)
	GetNamespaceById(ctx context.Context, id int) (*K8sNamespace, error)

	DeleteDeploymentById(ctx context.Context, deployid int) (bool, error)
	DeleteProjectById(ctx context.Context, proid int) (bool, error)
	DeleteRouteById(ctx context.Context, routeid int) (bool, error)
	DeleteServiceById(ctx context.Context, serviceid int) (bool, error)

	InsertIntoConfig(ctx context.Context, config Config) (*Config, error)
	GetDeploymentByProjectId(ctx context.Context, projectId int) ([]*Deployment, error)
	GetConfigByProjectIdAndNamespaceId(ctx context.Context, projectId int, namespaceId int, size int, page int) ([]*Config, error)
	GetConfigByProjectId(ctx context.Context, projectId int, size int, page int) ([]*Config, error)
	GetConfigsCountByProIdAndNspId(ctx context.Context, proid int, nspid int) (int, error)
	UpdateConfigById(ctx context.Context, config Config) (int, error)
	GetDeploymentByFuzzyFind(ctx context.Context, fuzzystr string, namespaceId int, projectId int) ([]*Deployment, error)
	SelectDingtalkBotByPro(ctx context.Context, projectId int) ([]*DingTalkBot, error)
	InsertIntoDingtalkBot(ctx context.Context, dingtalkBot DingTalkBot) (*DingTalkBot, error)
	GetConfigByDeploymentId(ctx context.Context, deploymentId int) ([]*Config, error)
	DeleteConfigById(ctx context.Context, id int) (bool, error)
	DeleteDingtalkBotById(ctx context.Context, id int) (bool, error)
}
type DbUtils interface {
}

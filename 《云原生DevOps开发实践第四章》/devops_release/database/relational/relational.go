package relational

import (
	"context"

	. "devops_release/database/model"
)

type DevopsDb interface {
	GetClusterById(ctx context.Context, id int) (*K8sClusterInfo, error)
	GetNspByClusterAndName(ctx context.Context, clusterId int, name string) (*K8sNamespace, error)
	GetAllPro(ctx context.Context) ([]*Project, error)
	GetProIdByName(ctx context.Context, name string) (*Project, error)
	GetClusterIdByName(ctx context.Context, name string) (*K8sClusterInfo, error)
	GetRoutesByProId(ctx context.Context, projectId int) ([]*Route, error)
	GetProjectByName(ctx context.Context, projectName string) (*Project, error)
	GetProjectById(ctx context.Context, projectId int) (*Project, error)
	GetRoutesByProIdAndNspId(ctx context.Context, projectId int, namespaceId int) ([]*Route, error)
	GetDeploymentsByProAndNamespace(ctx context.Context, projectId int, namespaceId int) ([]*Deployment, error)
	GetDeploymentByName(ctx context.Context, name string) ([]*Deployment, error)
	GetServicesByDeploymentId(ctx context.Context, deploymentId int) ([]*Service, error)
	GetConfigByDeploymentId(ctx context.Context, deploymentId int) ([]*Config, error)
	GetDeploymentById(ctx context.Context, deploymentId int) (*Deployment, error)
	GetConfigByNamespaceIdAndConfigName(ctx context.Context, namespaceId int, configName string) ([]*Config, error)
	GetConfigById(ctx context.Context, configId int) (*Config, error)
	SelectDingtalkBotByPro(ctx context.Context, projectId int) ([]*DingTalkBot, error)

	InsertIntoDeployment(ctx context.Context, deployment Deployment) (*Deployment, error)
	InsertIntoDockerInfo(ctx context.Context, dockerInfo DockerInfo) (*DockerInfo, error)
	InsertIntoK8sCluster(ctx context.Context, k8sClusterInfo K8sClusterInfo) (*K8sClusterInfo, error)
	InsertIntoRoutes(ctx context.Context, route Route) (*Route, error)
	InsertIntoK8sNamespace(ctx context.Context, nsp K8sNamespace) (*K8sNamespace, error)
	InsertIntoProjects(ctx context.Context, project Project) (*Project, error)
	InsertIntoService(ctx context.Context, service Service) (*Service, error)
	InsertIntoConfig(ctx context.Context,config Config) (*Config,error)
	GetWatcherByRouteAndUser(ctx context.Context, routeId int, userUuid string) (*Watcher, error)
	GetWatcherUUIDByClusterNameAndProId(ctx context.Context, projectId int, clusterName string) ([]*Watcher, error)
	GetWatcherUUIDByNspNameAndProId(ctx context.Context, projectId int, nspName string) ([]*Watcher, error)
	GetNamespaceById(ctx context.Context, namespaceId int) (*K8sNamespace, error)
	UpdateDeploymentContentById(ctx context.Context, deployment Deployment) (*Deployment, error)
}
type DbUtils interface {
}

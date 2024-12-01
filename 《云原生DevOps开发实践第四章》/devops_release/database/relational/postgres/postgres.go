package postgres

import (
	"context"

	. "devops_release/database/model"
	_ "github.com/lib/pq"
)

const (
	SELECT_DEPLOYMENTS_BY_PRO_AND_CHANNEL 		   = "select * from deployment where project_id=:project_id and channel_name=:channel_name"
	SELECT_DEPLOYMENTS_BY_NAME 					   = "select * from deployment where deployment_name=:deployment_name"
	SELECT_ROUTES_BY_PRO_AND_CHANNEL               = "select distinct project_id,channel,cluster_id,nsp_name,docker,enabled from routes where project_id=:project_id and channel=:channel"
	SELECT_ROUTES_BY_PRO_ID                        = "select distinct project_id,channel,cluster_id,nsp_name,docker,enabled from routes where project_id=:project_id"
	SELECT_CLUSTER_BY_ID                           = "select * from k8s_cluster where id=:id"
	SELECT_NSP_BY_CLUSTER_ID_AND_NAME              = "select * from k8s_namespace where k8s_cluster_id=:k8s_cluster_id and name=:name"
	SELECT_PROJECT_BY_ID                           = "select * from projects where id=:id"
	SELECT_PROJECT_ID_BY_NAME                      = "select id from projects where project_name=:project_name"
	SELECT_CLUSTER_ID_BY_NAME                      = "select id from k8s_cluster where name=:name"
	SELECT_SERVICES_BY_DEPLOYMENT_ID               = "select * from service where deployment_id=:deployment_id"
	SELECT_PROJECT_BY_NAME                         = "select * from projects where project_name=:project_name"
	SELECT_WATCHER_BY_ROUTER_AND_USER              = "select * from watcher where route_id =:route_id and user_uuid=:user_uuid"
	SELECT_WATCHER_UUID_BY_CLUSTER_NAME_AND_PRO_ID = "select user_uuid from watcher left join routes on route_id = routes.id left join k8s_cluster on cluster_id = k8s_cluster.id where k8s_cluster.name = :cluster_name and project_id=:project_id"
	SELECT_WATCHER_UUIDS_BY_NSP_NAME_AND_PRO_ID    = "select user_uuid from watcher left join routes on route_id = routes.id  where nsp_name=:nsp_name and project_id=:project_id"
	SELECT_NAMESPACE_BY_ID                         = "select * from k8s_namespace where id=:id"
	SELECT_ROUTES_BY_PRO_AND_NAMESPACE             = "select distinct project_id,cluster_id,nsp_name,namespace_id,docker,enabled from routes where project_id=:project_id and namespace_id=:namespace_id"
	SELECT_DEPLOYMENTS_BY_PRO_AND_NAMESPACE        = "select * from deployment where project_id=:project_id and namespace_id=:namespace_id"
	SELECT_APOLLO_CONFIG_BY_PRO_ID                 = "select * from apollo_config where project_id=:project_id"
	SELECT_APOLLO_CONFIG_BY_APOLLO_CONFIG_NAME     = "select * from apollo_config where apollo_config_name=:apollo_config_name"
	SELECT_DEPLOYMENT_BY_ID                        = "select * from deployment where id=:id"
	SELECT_DINGTALK_BOT_BY_PRO                     = "select * from dingtalk_bot where project_id=:project_id"

	SELECT_ALL_CONFIG                             = "select * from cmcfg where 1=1"
	SELECT_CONFIG_BY_DEPLOYMENT_ID                = "select * from cmcfg where deployment_id=:deployment_id"
	SELECT_CONFIG_BY_PROJECT_ID                   = "select * from cmcfg where project_id=:project_id"
	INSERT_INTO_CONFIG                            = "insert into cmcfg (project_id,file_name,configmap_name,restart_after_pub,content,namespace_id,config_name,deployment_id) values(:project_id,:file_name,:configmap_name,:restart_after_pub,:content,:namespace_id,:config_name,:deployment_id) returning id"
	UPDATE_CONFIG_BY_ID                           = "update cmcfg set project_id=:project_id,file_name=:file_name,configmap_name=:configmap_name,restart_after_pub=:restart_after_pub,content=:content,namespace_id=:namespace_id,config_name=:config_name,deployment_id=:deployment_id"
	SELECT_CONFIG_BY_NAMESPACE_ID_AND_CONFIG_NAME = "select * from cmcfg where namespace_id=:namespace_id and config_name=:config_name"
	SELECT_CONFIG_BY_ID                           = "select * from cmcfg where id=:id"

	INSERT_INTO_DEPLOYMENT    = "insert into deployment (deployment_name,project_id,content,enabled,docker_repo_id,channel_name,namespace_id) values(:deployment_name,:project_id,:content,:enabled,:docker_repo_id,:channel_name,:namespace_id) returning id"
	INSERT_INTO_DOCKER_INFO   = "insert into docker_info (name,type,username,password,registry_url,namespace) values(:name,:type,:username,:password,:registry_url,:namespace) returning id"
	INSERT_INTO_K8S_CLUSTER   = "insert into k8s_cluster (name,ca,token,url,ip) values(:name,:ca,:token,:url,:ip) returning id"
	INSERT_INTO_K8S_NAMESPACE = "insert into k8s_namespace (k8s_cluster_id,name,description,config) values(:k8s_cluster_id,:name,:description,:config) returning id"
	INSERT_INTO_ROUTES        = "insert into routes (project_id,ref_rep,cluster_id,nsp_name,docker,enabled,create_time,update_time) values(:project_id,:ref_rep,:channel,:cluster_id,:nsp_name,:docker,:enabled,:create_time,:update_time) returning id"
	INSERT_INTO_SERVICE       = "insert into service (deployment_id,content) values(:deployment_id,:content) returning id"
	INSERT_INTO_PROJECTS      = `insert into projects(repo_url,project_name,repo_type,tags,enabled,project_token,enabled_branchs) values(:repo_url,:project_name,:repo_type,:tags,:enabled,:project_token,:enabled_branchs) returning id,project_name`

	UPDATE_DEPLOYMENT_CONTENT_BY_Id = `update deployment set content =:content where id = :id returning project_id`
)

var DevopsDb = &DevOpsDbImpl{}

type DevOpsDbImpl struct {
}

// func (d *DevOpsDbImpl) GetDeploymentsByProAndChannel(ctx context.Context, projectId int, channel string) ([]*Deployment, error) {
// 	deployments := []*Deployment{}
// 	err := PostgresUtils.PrepareQuery(ctx, SELECT_DEPLOYMENTS_BY_PRO_AND_CHANNEL, &deployments, Deployment{ProjectId: projectId, ChannelName: channel})
// 	if err != nil {
// 		nlog.Errorf(err.Error())
// 		return nil, err
// 	}
// 	return deployments, nil
// }
// func (d *DevOpsDbImpl) GetRoutesByProAndChannel(ctx context.Context, projectId int, channel string) ([]*Route, error) {
// 	routes := []*Route{}
// 	err := PostgresUtils.PrepareQuery(ctx, SELECT_ROUTES_BY_PRO_AND_CHANNEL, &routes, Route{ProjectId: projectId, Channel: channel})
// 	if err != nil {
// 		return nil, err
// 	}
// 	return routes, nil
// }
func (d *DevOpsDbImpl) GetClusterById(ctx context.Context, id int) (*K8sClusterInfo, error) {
	k8sClusterInfo := &K8sClusterInfo{}
	err := PostgresUtils.PrepareQueryRaw(ctx, SELECT_CLUSTER_BY_ID, k8sClusterInfo, K8sClusterInfo{Id: id})
	if err != nil {
		return nil, err
	}
	return k8sClusterInfo, nil
}
func (d *DevOpsDbImpl) GetNspByClusterAndName(ctx context.Context, clusterId int, name string) (*K8sNamespace, error) {
	namespace := &K8sNamespace{}
	err := PostgresUtils.PrepareQueryRaw(ctx, SELECT_NSP_BY_CLUSTER_ID_AND_NAME, namespace, K8sNamespace{K8sClusterId: clusterId, Name: name})
	if err != nil {
		return nil, err
	}
	return namespace, nil
}
func (d *DevOpsDbImpl) GetAllPro(ctx context.Context) ([]*Project, error) {
	res := []*Project{}
	err := PostgresUtils.PrepareQuery(ctx, "select * from projects where 1=1", &res, Project{})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (d *DevOpsDbImpl) GetProIdByName(ctx context.Context, name string) (*Project, error) {
	res := &Project{}
	err := PostgresUtils.PrepareQueryRaw(ctx, SELECT_PROJECT_ID_BY_NAME, res, Project{ProjectName: name})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (d *DevOpsDbImpl) GetProjectByName(ctx context.Context, name string) (*Project, error) {
	res := &Project{}
	err := PostgresUtils.PrepareQueryRaw(ctx, SELECT_PROJECT_BY_NAME, res, Project{ProjectName: name})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (d *DevOpsDbImpl) GetProjectById(ctx context.Context, projectId int) (*Project, error) {
	res := &Project{}
	err := PostgresUtils.PrepareQueryRaw(ctx, SELECT_PROJECT_BY_ID, res, Project{Id: projectId})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (d *DevOpsDbImpl) GetClusterIdByName(ctx context.Context, name string) (*K8sClusterInfo, error) {
	res := &K8sClusterInfo{}
	err := PostgresUtils.PrepareQueryRaw(ctx, SELECT_CLUSTER_ID_BY_NAME, res, K8sClusterInfo{Name: name})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (d *DevOpsDbImpl) GetDeploymentsByProAndNamespace(ctx context.Context, projectId int, namespaceId int) ([]*Deployment, error) {
	res := []*Deployment{}
	err := PostgresUtils.PrepareQuery(ctx, SELECT_DEPLOYMENTS_BY_PRO_AND_NAMESPACE, &res, Deployment{ProjectId: projectId, NamespaceId: namespaceId})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (d *DevOpsDbImpl) GetDeploymentByName(ctx context.Context, name string) ([]*Deployment, error) {
	res := []*Deployment{}
	err := PostgresUtils.PrepareQuery(ctx, SELECT_DEPLOYMENTS_BY_NAME, &res, Deployment{DeploymentName: name})
	return res, err
}

func (d *DevOpsDbImpl) GetServicesByDeploymentId(ctx context.Context, deploymentId int) ([]*Service, error) {
	res := []*Service{}
	err := PostgresUtils.PrepareQuery(ctx, SELECT_SERVICES_BY_DEPLOYMENT_ID, &res, Service{DeploymentId: deploymentId})
	return res, err
}

// func (d *DevOpsDbImpl) GetServicesByDeploymentIdAndChannel(ctx context.Context, deploymentId int, channel string) ([]*Service, error) {
// 	res := []*Service{}
// 	err := PostgresUtils.PrepareQuery(ctx, SELECT_SERVICES_BY_DEPLOYMENT_ID_AND_CHANNEL, &res, Service{DeploymentId: deploymentId, ChannelName: channel})
// 	if err != nil {
// 		return nil, err
// 	}
// 	return res, nil
// }

func (d *DevOpsDbImpl) GetRoutesByProId(ctx context.Context, projectId int) ([]*Route, error) {
	res := []*Route{}
	err := PostgresUtils.PrepareQuery(ctx, SELECT_ROUTES_BY_PRO_ID, &res, Route{ProjectId: projectId})
	if err != nil {
		return nil, err
	}
	return res, nil
}
func (d *DevOpsDbImpl) GetWatcherByRouteAndUser(ctx context.Context, routeId int, userUuid string) (*Watcher, error) {
	res := &Watcher{}
	err := PostgresUtils.PrepareQueryRaw(ctx, SELECT_WATCHER_BY_ROUTER_AND_USER, &res, Watcher{RouteId: routeId, UserUUId: userUuid})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (d *DevOpsDbImpl) GetRoutesByProIdAndNspId(ctx context.Context, projectId int, namespaceId int) ([]*Route, error) {
	res := []*Route{}
	err := PostgresUtils.PrepareQuery(ctx, SELECT_ROUTES_BY_PRO_AND_NAMESPACE, &res, Route{ProjectId: projectId, NamespaceId: namespaceId})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (d *DevOpsDbImpl) GetWatcherUUIDByClusterNameAndProId(ctx context.Context, projectId int, clusterName string) ([]*Watcher, error) {
	res := []*Watcher{}
	err := PostgresUtils.PrepareQuery(ctx, SELECT_WATCHER_UUID_BY_CLUSTER_NAME_AND_PRO_ID, res, struct {
		ClusterName string `db:"cluster_name"`
		ProjectId   int    `db:"project_id"`
	}{
		ClusterName: clusterName,
		ProjectId:   projectId,
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}
func (d *DevOpsDbImpl) GetNamespaceById(ctx context.Context, namespaceId int) (*K8sNamespace, error) {
	res := &K8sNamespace{}
	err := PostgresUtils.PrepareQueryRaw(ctx, SELECT_NAMESPACE_BY_ID, res, K8sNamespace{Id: namespaceId})
	if err != nil {
		return nil, err
	}
	return res, nil
}
func (d *DevOpsDbImpl) GetWatcherUUIDByNspNameAndProId(ctx context.Context, projectId int, nspName string) ([]*Watcher, error) {
	res := []*Watcher{}
	err := PostgresUtils.PrepareQuery(ctx, SELECT_WATCHER_UUID_BY_CLUSTER_NAME_AND_PRO_ID, &res, struct {
		NspName   string `db:"nsp_name"`
		ProjectId int    `db:"project_id"`
	}{
		NspName:   nspName,
		ProjectId: projectId,
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

// INSERT_INTO_DEPLOYMENT    = "insert into deployment (deployment_name,project_id,channel_name,content,enabled,docker_repo_id) values(?,?,?,?,?,?)"
// INSERT_INTO_DOCKER_INFO   = "insert into docker_info (name,type,username,password,registry_url,namespace) values(?,?,?,?,?,?)"
// INSERT_INTO_K8S_CLUSTER   = "insert into k8s_cluster (name,ca,token,url,ip) values(?,?,?,?,?)"

// INSERT_INTO_ROUTES        = "insert into routes (project_id,ref_rep,channel,cluster_id,nsp_name,docker,create_time,update_time)"
func (d *DevOpsDbImpl) InsertIntoDeployment(ctx context.Context, deployment Deployment) (*Deployment, error) {
	res := &Deployment{}
	err := PostgresUtils.PrepareQueryRaw(ctx, INSERT_INTO_DEPLOYMENT, res, deployment)
	if err != nil {
		return nil, err
	}
	return res, nil
}
func (d *DevOpsDbImpl) InsertIntoDockerInfo(ctx context.Context, dockerInfo DockerInfo) (*DockerInfo, error) {
	res := &DockerInfo{}
	err := PostgresUtils.PrepareQueryRaw(ctx, INSERT_INTO_DOCKER_INFO, res, dockerInfo)
	if err != nil {
		return nil, err
	}
	return res, nil
}
func (d *DevOpsDbImpl) InsertIntoK8sCluster(ctx context.Context, k8sClusterInfo K8sClusterInfo) (*K8sClusterInfo, error) {
	res := &K8sClusterInfo{}
	err := PostgresUtils.PrepareQueryRaw(ctx, INSERT_INTO_K8S_CLUSTER, res, k8sClusterInfo)
	if err != nil {
		return nil, err
	}
	return res, nil
}
func (d *DevOpsDbImpl) InsertIntoRoutes(ctx context.Context, route Route) (*Route, error) {
	res := &Route{}
	err := PostgresUtils.PrepareQueryRaw(ctx, INSERT_INTO_ROUTES, res, route)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// INSERT_INTO_K8S_NAMESPACE = "insert into k8s_namespace (k8s_cluster_id,name,description,config) values(?,?,?,?)"
func (d *DevOpsDbImpl) InsertIntoK8sNamespace(ctx context.Context, nsp K8sNamespace) (*K8sNamespace, error) {
	res := &K8sNamespace{}
	err := PostgresUtils.PrepareQueryRaw(ctx, INSERT_INTO_K8S_NAMESPACE, res, nsp)
	if err != nil {
		return nil, err
	}
	return res, nil
}

//INSERT_INTO_PROJECTS      = "insert into projects (repo_url,project_name,repo_type,tags,enabled,project_token)"
func (d *DevOpsDbImpl) InsertIntoProjects(ctx context.Context, project Project) (*Project, error) {
	res := &Project{}
	err := PostgresUtils.PrepareQueryRaw(ctx, INSERT_INTO_PROJECTS, res, project)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (d *DevOpsDbImpl) InsertIntoService(ctx context.Context, service Service) (*Service, error) {
	res := &Service{}
	err := PostgresUtils.PrepareQueryRaw(ctx, INSERT_INTO_SERVICE, res, service)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (d *DevOpsDbImpl) InsertIntoConfig(ctx context.Context, config Config) (*Config, error) {
	res := &Config{}
	err := PostgresUtils.PrepareQueryRaw(ctx, INSERT_INTO_CONFIG, res, config)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (d *DevOpsDbImpl) UpdateDeploymentContentById(ctx context.Context, deployment Deployment) (*Deployment, error) {
	res := &Deployment{}
	err := PostgresUtils.PrepareQueryRaw(ctx, UPDATE_DEPLOYMENT_CONTENT_BY_Id, res, deployment)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (d *DevOpsDbImpl) GetConfigByDeploymentId(ctx context.Context, deploymentId int) ([]*Config, error) {
	res := []*Config{}
	err := PostgresUtils.PrepareQuery(ctx, SELECT_CONFIG_BY_DEPLOYMENT_ID, &res, Config{DeploymentId: deploymentId})
	return res, err
}

func (d *DevOpsDbImpl) GetDeploymentById(ctx context.Context, deploymentId int) (*Deployment, error) {
	res := &Deployment{}
	err := PostgresUtils.PrepareQueryRaw(ctx, SELECT_DEPLOYMENT_BY_ID, res, Deployment{Id: deploymentId})
	return res, err
}

func (d *DevOpsDbImpl) GetConfigByNamespaceIdAndConfigName(ctx context.Context, namespaceId int, configName string) ([]*Config, error) {
	res := []*Config{}
	err := PostgresUtils.PrepareQuery(ctx, SELECT_CONFIG_BY_NAMESPACE_ID_AND_CONFIG_NAME, &res, Config{NamespaceId: namespaceId, ConfigName: configName})
	return res, err
}

func (d *DevOpsDbImpl) GetConfigById(ctx context.Context, configId int) (*Config, error) {
	res := &Config{}
	err := PostgresUtils.PrepareQueryRaw(ctx, SELECT_CONFIG_BY_ID, res, Config{Id: configId})
	return res, err
}

func (d *DevOpsDbImpl) SelectDingtalkBotByPro(ctx context.Context, projectId int) ([]*DingTalkBot, error) {
	res := []*DingTalkBot{}
	err := PostgresUtils.PrepareQuery(ctx, SELECT_DINGTALK_BOT_BY_PRO, &res, DingTalkBot{ProjectId: projectId})
	return res, err
}

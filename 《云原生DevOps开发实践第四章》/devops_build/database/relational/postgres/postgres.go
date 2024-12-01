package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	. "devops_build/database/model"
	//"time"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

const (
	SELECT_DEPLOYMENTS_BY_PRO_AND_CHANNEL = "select * from deployment where project_id=:project_id and channel_name=:channel_name"
	SELECT_DEPLOYMENTS_BY_PRO_AND_NSP     = "select * from deployment where project_id=:project_id and namespace_id=:namespace_id order by id limit :size offset :count "

	SELECT_DEPLOYMENTS_COUNT_BY_PRO_AND_NSP = "select count(*) as count from deployment where project_id=:project_id and namespace_id=:namespace_id"

	SELECT_ROUTES_BY_PRO_AND_NSP       = "select * from routes where project_id=:project_id and namespace_id=:namespace_id order by id limit :size offset :count "
	SELECT_ROUTES_COUNT_BY_PRO_AND_NSP = "select count(*) as count from routes where project_id=:project_id and namespace_id=:namespace_id"

	SELECT_SERVICES_BY_DEPID       = "select * from service where deployment_id=:deployment_id"
	SELECT_SERVICES_COUNT_BY_DEPID = "select count(*) as count from service where deployment_id=:deployment_id"

	SELECT_ROUTES_BY_PRO_AND_CHANNEL                         = "select distinct project_id,channel,cluster_id,nsp_name,docker,enabled from routes where project_id=:project_id and channel=:channel" //nolint:lll
	SELECT_ROUTES_BY_PRO_ID                                  = "select distinct project_id,channel,cluster_id,nsp_name,docker,enabled,namespace_id from routes where project_id=:project_id"         //nolint:lll
	SELECT_ROUTES_REP_NAMESPACE_BY_PROJECT_ID                = "select distinct ref_rep,namespace_id from routes where project_id=:project_id"
	SELECT_ALL_ROUTES                                        = "SELECT * from routes where 1=1"
	SELECT_CLUSTER_BY_ID                                     = "select * from k8s_cluster where id=:id"
	SELECT_NSP_BY_CLUSTER_ID_AND_NAME                        = "select * from k8s_namespace where k8s_cluster_id=:k8s_cluster_id and name=:name"
	SELECT_PROJECT_BY_ID                                     = "select * from projects where id=:id"
	SELECT_PROJECT_ID_BY_NAME                                = "select id from projects where project_name=:project_name"
	SELECT_CLUSTER_ID_BY_NAME                                = "select id from k8s_cluster where name=:name"
	SELECT_SERVICES_BY_DEPLOYMENT_ID_AND_CHANNEL             = "select * from service where deployment_id=:deployment_id and channel_name=:channel_name"
	SELECT_PROJECT_BY_NAME                                   = "select * from projects where project_name=:project_name"
	SELECT_PROJECT_ENABLED_BRANCHS_AND_TOKEN_BY_PROJECT_NAME = "select distinct id,enabled_branchs,project_token from projects where project_name=:project_name"
	SELECT_ROUTE_IDS_BY_KEYWORDS                             = "select id from routes where ${condition}"

	SELECT_DEPLOYMENT_COUNT = "select count(*) as count from deployment"
	SELECT_ROUTE_COUNT      = "select count(*) as count from routes"
	SELECT_PROJECT_COUNT    = "select count(*) as count from projects"
	SELECT_SERVICE_COUNT    = "select count(*) as count from service"

	SELECT_PROJECTS_BY_PAGESIZE        = "select * from projects where id >= (select id from projects order by id limit 1 offset :count) order by id limit :size"
	SELECT_DEPLOYMENTS_BY_PAGESIZE     = "select * from deployment where id >= (select id from deployment order by id limit 1 offset :count) order by id limit :size"
	SELECT_SERVICES_BY_PAGESIZE        = "select * from service where id >= (select id from service order by id limit 1 offset :count) order by id limit :size"
	SELECT_ROUTES_BY_PAGESIZE          = "select * from routes where id >= (select id from service order by id limit 1 offset :count) order by id limit :size"
	SELECT_PROJECTS_BY_PAGE            = "select * from projects where id >= (select id from projects order by id limit 1 offset :count) order by id limit 20"
	SELECT_PROJECT_BY_FUZZY            = "select * from projects where project_name ~ :fuzzystr"
	SELECT_DEPLOYMENTS_BY_PAGE         = "select * from deployment where id >= (select id from deployment order by id limit 1 offset :count) order by id limit 20"
	SELECT_SERVICES_BY_PAGE            = "select * from service where id >= (select id from service order by id limit 1 offset :count) order by id limit 20"
	SELECT_ROUTES_BY_PAGE              = "select * from routes where id >= (select id from service order by id limit 1 offset :count) order by id limit 20"
	SELECT_REQUEST_HISTORY_BY_PAGESIZE = "select * from request_history where id >= (select id from request_history order by id limit 1 offset :count) order by id limit :size"

	SELECT_DEPLOYMENT_BY_ID         = "select * from deployment where id=:id"
	SELECT_DEPLOYMENT_BY_PROJECT_ID = "select * from deployment where project_id=:project_id"
	SELECT_OPS_HISTORY_BY_PAGESIZE  = "select * from ops_history where id >= (select id from ops_history order by id limit 1 offset :count) order by id limit :size"

	SELECT_DEPLOYMENTS_BY_NAME = "select * from deployment where deployment_name=:deployment_name"
	//EXIST_DEPLOYMENT_BY_NAME_AND_PRO_AND_CHANNEL = "select * from deployment where project_id=:project_id and channel_name=:channel_name and deployment_name=:deployment_name limit 1"
	////EXIST_PROJECT_BY_NAME              = "select top 1 project_name from projects where project_name=:project_name"
	////EXIST_SERVICE_BY__AND_CHANNEL    = "select top 1 id from service where deployment_id=:deployment_id and channel_name=:channel_name"
	//EXIST_ROUTE_BY_PRO_AND_CHANNEL_AND_REP_AND_NSP_AND_CLUSTER = "select * from routes where project_id=:project_id and ref_rep=:ref_rep and channel=:channel and nsp_name=:nsp_name and cluster_id=:cluster_id limit 1"

	SELECT_ROUTE_BY_PRO_AND_CHANNEL_AND_REP_AND_NSP_AND_CLUSTER = "select * from routes where project_id=:project_id and ref_rep=:ref_rep and channel=:channel and nsp_name=:nsp_name and cluster_id=:cluster_id"
	SELECT_DEPLYMENT_BY_NAME_AND_PRO_AND_CHANNEL                = "select * from deployment where project_id=:project_id and channel_name=:channel_name and deployment_name=:deployment_name"
	SELECT_PROJECT_BY_GITLAB_ID                                 = "select * from projects where gitlab_id=:gitlab_id"
	SELECT_NAMESPACE_BY_ID                                      = "select * from k8s_namespace where id=:id"

	SELECT_ALL_CONFIG                         = "select * from cmcfg where 1=1"
	SELECT_CONFIG_BY_DEPLOYMENT_ID            = "select * from cmcfg where deployment_id=:deployment_id"
	SELECT_CONFIG_BY_PROJECT_ID               = "select * from cmcfg where project_id=:project_id"
	INSERT_INTO_CONFIG                        = "insert into cmcfg (project_id,file_name,configmap_name,restart_after_pub,content,namespace_id,config_name,deployment_id) values(:project_id,:file_name,:configmap_name,:restart_after_pub,:content,:namespace_id,:config_name,:deployment_id) returning id"
	UPDATE_CONFIG_BY_ID                       = "update cmcfg set project_id=:project_id,file_name=:file_name,configmap_name=:configmap_name,restart_after_pub=:restart_after_pub,content=:content,namespace_id=:namespace_id,config_name=:config_name,deployment_id=:deployment_id where id=:id"
	SELECT_CONFIGS_BY_PRO_ID                  = "select * from cmcfg where project_id=:project_id  order by deployment_id limit :size offset :count "
	SELECT_CONFIGS_BY_PRO_ID_AND_NAMESPACE_ID = "select * from cmcfg where project_id=:project_id and namespace_id=:namespace_id  order by deployment_id limit :size offset :count"
	SELECT_CONFIGS_COUNT_BY_PRO_AND_NSP       = "select count(*) as count from cmcfg where project_id=:project_id and namespace_id=:namespace_id"
	SELECT_CONFIGS_COUNT_BY_PRO               = "select count(*) as count from cmcfg where project_id=:project_id"

	INSERT_INTO_DEPLOYMENT      = "insert into deployment (deployment_name,project_id,channel_name,content,enabled,docker_repo_id,namespace_id) values(:deployment_name,:project_id,:channel_name,:content,:enabled,:docker_repo_id,:namespace_id) returning id" //nolint:lll
	INSERT_INTO_DOCKER_INFO     = "insert into docker_info (name,type,username,password,registry_url,namespace) values(:name,:type,:username,:password,:registry_url,:namespace) returning id"                                                                   //nolint:lll
	INSERT_INTO_K8S_CLUSTER     = "insert into k8s_cluster (name,ca,token,url,ip) values(:name,:ca,:token,:url,:ip) returning id"
	INSERT_INTO_K8S_NAMESPACE   = "insert into k8s_namespace (k8s_cluster_id,name,description,config) values(:k8s_cluster_id,:name,:description,:config) returning id"                                                                                                                        //nolint:lll
	INSERT_INTO_ROUTES          = "insert into routes (project_id,ref_rep,channel,cluster_id,nsp_name,docker,enabled,create_time,update_time,namespace_id) values(:project_id,:ref_rep,:channel,:cluster_id,:nsp_name,:docker,:enabled,:create_time,:update_time,:namespace_id) returning id" //nolint:lll
	INSERT_INTO_SERVICE         = "insert into service (deployment_id,channel_name,content) values(:deployment_id,:channel_name,:content) returning id"
	INSERT_INTO_PROJECTS        = `insert into projects(repo_url,project_name,repo_type,tags,enabled,project_token,enabled_branchs,descript,"group",topic,gitlab_id) values(:repo_url,:project_name,:repo_type,:tags,:enabled,:project_token,:enabled_branchs,:descript,:group,:topic,:gitlab_id) returning id,project_name`
	INSERT_INTO_REQUEST_HISTORY = "insert into request_history(request_url,user_id,request_params,request_method,host,update_time) values(:request_url,:user_id,:request_params,:request_method,:host,:update_time) returning id"
	INSERT_INTO_OPS_HISTORY     = "insert into ops_history(resource_type,ops_type,resource_id,update_time,change_before,change_after,user_id) values(:resource_type,:ops_type,:resource_id,:update_time,:change_before,:change_after,:user_id) returning id"
	INSERT_INTO_APOLLO_CONFIG   = "insert into apollo_config(project_id,apollo_config_name,apollo_token,apollo_secret,file_name,configmap_name) values(:project_id,:apollo_config_name,:apollo_token,:apollo_secret,:file_name,:configmap_name) returning id"

	UPDATE_CLUSTER_INFO     = "update k8s_cluster set name=:name,ca=:ca,token=:token,url=:url,ip=:ip,encrypt_ca=:encrypt_ca,encrypt_token=:encrypt_token where id=:id"
	UPDATE_DEPLOYMENT_BY_ID = "update deployment set deployment_name=:deployment_name,project_id=:project_id,channel_name=:channel_name,content=:content,enabled=:enabled,docker_repo_id=:docker_repo_id,namespace_id=:namespace_id where id=:id"
	UPDATE_PROJECT_BY_ID    = `update projects set repo_url=:repo_url,project_name=:project_name,repo_type=:repo_type,tags=:tags,enabled=:enabled,project_token=:project_token,enabled_branchs=:enabled_branchs,descript=:descript,"group"=:group,gitlab_id=:gitlab_id,topic=:topic  where id=:id`
	UPDATE_SERVICE_BY_ID    = "update service set deployment_id=:deployment_id,channel_name=:channel_name,content=:content where id=:id"
	SELECT_USER_BY_NAME     = "select * from users where name=:name"
	SELECT_USER_BY_ACCOUNT  = "select * from users where account=:account"
	INSERT_INTO_USER        = "insert into users (name,account,password,department,tel,admin) values(:name,:account,:password,:department,:tel,:admin) returning uuid,name,account"

	SELECT_CLUSTERS         = "select * from k8s_cluster"
	SELECT_NSP_BY_CLUSTERID = "select * from k8s_namespace where k8s_cluster_id=:k8s_cluster_id"
	UPDATE_ROUTE_BY_ID      = "update routes set project_id=:project_id ,ref_rep=:ref_rep,channel=:channel,nsp_name=:nsp_name,docker=:docker,update_time=:update_time,cluster_id=:cluster_id,enabled=:enabled,namespace_id=:namespace_id where id=:id"

	DELETE_DEPLOYMENT_BY_ID = "delete from deployment where id=:id"
	DELETE_ROUTE_BY_ID      = "delete from routes where id=:id"
	DELETE_PROJECT_BY_ID    = "delete from projects where id=:id"
	DELETE_SERVICE_BY_ID    = "delete from service where id=:id"

	SELECT_DEPLOYMENT_BY_FUZZY_PRO_AND_NAMESPACE = "select * from deployment where deployment_name ~ :fuzzystr and project_id=:project_id and namespace_id=:namespace_id"
	SELECT_DINGTALK_BOT_BY_PRO                   = "select * from dingtalk_bot where project_id=:project_id"
	INSERT_INTO_DINGTALK_BOT                     = "insert into dingtalk_bot (project_id,dingtalk_bot_hook,descript) values(:project_id,:dingtalk_bot_hook,:descript) returning id"
	UPDATE_DINGTALK_BOT                          = "update dingtalk_bot set project_id=:project_id,descript=:descript where id=:id"
	DELETE_DINGTALK_BOT                          = "delete from dingtalk_bot where id=:id"

	SELECT_CONFIGS_BY_DEPLOYMENT_ID = "select * from cmcfg where deployment_id=:deployment_id"

	DELETE_CONFIG_BY_CONFIG_ID = "delete from cmcfg where id=:id"
)

var DevopsDb = &DevOpsDbImpl{}

type DevOpsDbImpl struct {
}

//use by page search
type Count struct {
	Count int `db:"count"`
	Size  int `db:"size"`
}

func (d *DevOpsDbImpl) DeleteDeploymentById(ctx context.Context, deployid int) (bool, error) {
	res, err := PostgresUtils.PrepareExec(ctx, DELETE_DEPLOYMENT_BY_ID, Deployment{Id: deployid})
	if err != nil {
		return false, err
	}
	row, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	if row == 1 {
		return true, nil
	} else if row == 0 {
		return false, errors.New("can't find the deployment")
	} else {
		return false, errors.New(fmt.Sprintf("delete %d records of deployment", row))
	}
}

func (d *DevOpsDbImpl) DeleteRouteById(ctx context.Context, routeid int) (bool, error) {
	res, err := PostgresUtils.PrepareExec(ctx, DELETE_ROUTE_BY_ID, Route{Id: routeid})
	if err != nil {
		return false, err
	}
	row, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	if row == 1 {
		return true, nil
	} else if row == 0 {
		return false, errors.New("can't find the route")
	} else {
		return false, errors.New(fmt.Sprintf("delete %d records of route", row))
	}
}

func (d *DevOpsDbImpl) DeleteProjectById(ctx context.Context, proid int) (bool, error) {
	res, err := PostgresUtils.PrepareExec(ctx, DELETE_PROJECT_BY_ID, Project{Id: proid})
	if err != nil {
		return false, err
	}
	row, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	if row == 1 {
		return true, nil
	} else if row == 0 {
		return false, errors.New("can't find the project")
	} else {
		return false, errors.New(fmt.Sprintf("delete %d records of project", row))
	}
}

func (d *DevOpsDbImpl) DeleteServiceById(ctx context.Context, serviceid int) (bool, error) {
	res, err := PostgresUtils.PrepareExec(ctx, DELETE_SERVICE_BY_ID, Service{Id: serviceid})
	if err != nil {
		return false, err
	}
	row, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	if row == 1 {
		return true, nil
	} else if row == 0 {
		return false, errors.New("can't find the service")
	} else {
		return false, errors.New(fmt.Sprintf("delete %d records of service", row))
	}
}

func (d *DevOpsDbImpl) GetCluster(ctx context.Context) ([]*K8sClusterInfo, error) {
	res := []*K8sClusterInfo{}
	err := PostgresUtils.PrepareQuery(ctx, SELECT_CLUSTERS, &res, map[string]interface{}{})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (d *DevOpsDbImpl) GetNspByClusterId(ctx context.Context, clusterid int) ([]*K8sNamespace, error) {
	res := []*K8sNamespace{}
	err := PostgresUtils.PrepareQuery(ctx, SELECT_NSP_BY_CLUSTERID, &res, K8sNamespace{K8sClusterId: clusterid})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (d *DevOpsDbImpl) GetDeploymentsByProIdAndNspIdAndPageSize(ctx context.Context, proid int, nspid int, page int, size int) ([]*Deployment, error) {
	selectSql := SELECT_DEPLOYMENTS_BY_PRO_AND_NSP
	if nspid <= 0 {
		selectSql = strings.ReplaceAll(SELECT_DEPLOYMENTS_BY_PRO_AND_NSP, "and namespace_id=:namespace_id", "")
	}
	res := []*Deployment{}
	err := PostgresUtils.PrepareQuery(ctx, selectSql, &res, map[string]interface{}{"project_id": proid, "namespace_id": nspid, "count": (page - 1) * size, "size": size})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (d *DevOpsDbImpl) GetDeploymentsCountByProIdAndNspId(ctx context.Context, proid int, nspid int) (int, error) {
	selectSql := SELECT_DEPLOYMENTS_COUNT_BY_PRO_AND_NSP
	if nspid == 0 {
		selectSql = strings.ReplaceAll(SELECT_DEPLOYMENTS_COUNT_BY_PRO_AND_NSP, "and namespace_id=:namespace_id", "")
	}
	c := Count{}
	err := PostgresUtils.PrepareQueryRow(ctx, selectSql, &c, map[string]interface{}{"project_id": proid, "namespace_id": nspid})
	if err != nil {
		return 0, err
	}
	return c.Count, nil
}

func (d *DevOpsDbImpl) GetRoutesByProIdAndNspIdAndPageSize(ctx context.Context, proid int, nspid int, page int, size int) ([]*Route, error) {
	selectSql := SELECT_ROUTES_BY_PRO_AND_NSP
	if nspid <= 0 {
		selectSql = strings.ReplaceAll(SELECT_ROUTES_BY_PRO_AND_NSP, "and namespace_id=:namespace_id", "")
	}
	res := []*Route{}
	err := PostgresUtils.PrepareQuery(ctx, selectSql, &res, map[string]interface{}{"project_id": proid, "namespace_id": nspid, "count": (page - 1) * size, "size": size})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (d *DevOpsDbImpl) GetRoutesCountByProIdAndNspId(ctx context.Context, proid int, nspid int) (int, error) {
	selectSql := SELECT_ROUTES_COUNT_BY_PRO_AND_NSP
	if nspid <= 0 {
		selectSql = strings.ReplaceAll(SELECT_ROUTES_COUNT_BY_PRO_AND_NSP, "and namespace_id=:namespace_id", "")
	}
	c := Count{}
	err := PostgresUtils.PrepareQueryRow(ctx, selectSql, &c, map[string]interface{}{"project_id": proid, "namespace_id": nspid})
	if err != nil {
		return 0, err
	}
	return c.Count, nil
}

func (d *DevOpsDbImpl) GetDeploymentById(ctx context.Context, deploymentId int) (*Deployment, error) {
	res := &Deployment{}
	err := PostgresUtils.PrepareQueryRow(ctx, SELECT_DEPLOYMENT_BY_ID, res, Deployment{Id: deploymentId})
	return res, err
}

func (d *DevOpsDbImpl) GetServicesByDeploymentId(ctx context.Context, deployid int) ([]*Service, error) {
	res := []*Service{}
	err := PostgresUtils.PrepareQuery(ctx, SELECT_SERVICES_BY_DEPID, &res, map[string]interface{}{"deployment_id": deployid})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (d *DevOpsDbImpl) GetServicesCountByDeploymentId(ctx context.Context, deployid int) (int, error) {
	c := Count{}
	err := PostgresUtils.PrepareQueryRow(ctx, SELECT_SERVICES_COUNT_BY_DEPID, &c, map[string]interface{}{"deployment_id": deployid})
	if err != nil {
		return 0, err
	}
	return c.Count, nil
}

//Fuzzy find project
type FuzzyData struct {
	Fuzzystr string `db:"fuzzystr"`
}
type PageData struct {
	Count    int         `json:"count"`
	ListData interface{} `json:"list_data"`
}

func (d *DevOpsDbImpl) GetProjectById(ctx context.Context, projectId int) (*Project, error) {
	res := &Project{}
	err := PostgresUtils.PrepareQueryRow(ctx, SELECT_PROJECT_BY_ID, res, Project{Id: projectId})
	return res, err
}

func (d *DevOpsDbImpl) GetProjectCount(ctx context.Context) (int, error) {
	c := Count{}
	err := PostgresUtils.PrepareQueryRow(ctx, SELECT_PROJECT_COUNT, &c, Project{})
	if err != nil {
		return 0, err
	}
	return c.Count, nil
}

func (d *DevOpsDbImpl) GetServiceCount(ctx context.Context) (int, error) {
	c := Count{}
	err := PostgresUtils.PrepareQueryRow(ctx, SELECT_SERVICE_COUNT, &c, Service{})
	if err != nil {
		return 0, err
	}
	return c.Count, nil
}

func (d *DevOpsDbImpl) GetAllRoutes(ctx context.Context) ([]*Route, error) {
	res := []*Route{}
	err := PostgresUtils.PrepareQuery(ctx, SELECT_ALL_ROUTES, &res, Route{})
	return res, err
}

func (d *DevOpsDbImpl) GetRouteCount(ctx context.Context) (int, error) {
	c := Count{}
	err := PostgresUtils.PrepareQueryRow(ctx, SELECT_ROUTE_COUNT, &c, Route{})
	if err != nil {
		return 0, err
	}
	return c.Count, nil
}

func (d *DevOpsDbImpl) GetDeploymentCount(ctx context.Context) (int, error) {
	c := Count{}
	err := PostgresUtils.PrepareQueryRow(ctx, SELECT_DEPLOYMENT_COUNT, &c, Deployment{})
	if err != nil {
		return 0, err
	}
	return c.Count, nil
}

func (d *DevOpsDbImpl) IfExistProject(ctx context.Context, project Project) (bool, error) {
	_, err := d.GetProjectByGitlabId(ctx, project.GitlabId)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (d *DevOpsDbImpl) IfExistService(ctx context.Context, service Service) (bool, error) {
	_, err := d.GetServicesByDeploymentIdAndChannel(ctx, service.DeploymentId, service.ChannelName)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (d *DevOpsDbImpl) IfExistRoute(ctx context.Context, route Route) (bool, error) {
	//r := &Route{}
	//err := PostgresUtils.PrepareQueryRow(ctx, EXIST_ROUTE_BY_PRO_AND_CHANNEL_AND_REP_AND_NSP_AND_CLUSTER, r, route)
	//if errors.Is(err, sql.ErrNoRows) {
	//	return false, nil
	//}
	//if err != nil {
	//	return false, err
	//}
	//return true, nil
	_, err := d.GetRouteByProIdAndChannelAndRepAndNspAndCluster(ctx, route.ProjectId, route.Channel, route.RefRep, route.NspName, route.ClusterId)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (d *DevOpsDbImpl) IfExistDeployment(ctx context.Context, deployment Deployment) (bool, error) {
	_, err := d.GetDeploymentByNameAndProAndChannel(ctx, deployment.DeploymentName, deployment.ProjectId, deployment.ChannelName)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (d *DevOpsDbImpl) UpdateDeployment(ctx context.Context, deployment Deployment) (bool, error) {
	//not have this row:the row name is update
	//hav this row	   :the row other part update
	//have this row but id is not the project.id : this is conflict
	res, err := PostgresUtils.PrepareExec(ctx, UPDATE_DEPLOYMENT_BY_ID, deployment)
	if err != nil {
		return false, err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	if rows != 1 {
		return false, errors.New(fmt.Sprintf("expected to affect 1 row, affected %d", rows))
	}
	return true, nil
}

func (d *DevOpsDbImpl) UpdateRouteById(ctx context.Context, route Route) (int, error) {
	res, err := PostgresUtils.PrepareExec(ctx, UPDATE_ROUTE_BY_ID, route)
	if err != nil {
		return 0, err
	}
	fmt.Println("namespace:", route.NamespaceId)
	rowsEffect, err := res.RowsAffected()

	return int(rowsEffect), err
}

func (d *DevOpsDbImpl) UpdateRoute(ctx context.Context, route Route) (bool, error) {
	ro, err := d.GetRouteByProIdAndChannelAndRepAndNspAndCluster(ctx, route.ProjectId, route.Channel, route.RefRep, route.NspName, route.ClusterId)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false, err
	}
	//not have this row:the row name is update
	//hav this row	   :the row other part update
	//have this row but id is not the project.id : this is conflict
	if errors.Is(err, sql.ErrNoRows) || ro.Id == route.Id {
		res, err := PostgresUtils.PrepareExec(ctx, UPDATE_ROUTE_BY_ID, route)
		if err != nil {
			return false, err
		}
		rows, err := res.RowsAffected()
		if err != nil {
			return false, err
		}
		if rows != 1 {
			return false, errors.New(fmt.Sprintf("expected to affect 1 row, affected %d", rows))
		}
		return true, nil
	}
	return false, errors.New("update conflict")
}

func (d *DevOpsDbImpl) UpdateProject(ctx context.Context, project Project) (bool, error) {
	//not have this row:the row name is update
	//hav this row	   :the row other part update
	//have this row but id is not the project.id : this is conflict
	_, err := PostgresUtils.PrepareExec(ctx, UPDATE_PROJECT_BY_ID, project)
	if err != nil {
		return false, err
	}
	//_, err := res.RowsAffected()
	//if err != nil {
	//	return false, err
	//}
	////if rows != 1 {
	////	return false, errors.New(fmt.Sprintf("expected to affect 1 row, affected %d", rows))
	////}
	return true, nil
}

func (d *DevOpsDbImpl) UpdateService(ctx context.Context, service Service) (bool, error) {
	res, err := PostgresUtils.PrepareExec(ctx, UPDATE_SERVICE_BY_ID, service)
	if err != nil {
		return false, err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	if rows != 1 {
		return false, errors.New(fmt.Sprintf("expected to affect 1 row, affected %d", rows))
	}
	return true, nil
}

func (d *DevOpsDbImpl) GetDeploymentsByPageSize(ctx context.Context, page int, size int) ([]*Deployment, error) {
	res := []*Deployment{}
	err := PostgresUtils.PrepareQuery(ctx, SELECT_DEPLOYMENTS_BY_PAGESIZE, &res, Count{Count: (page - 1) * size, Size: size})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (d *DevOpsDbImpl) GetRoutesByPageSize(ctx context.Context, page int, size int) ([]*Route, error) {
	res := []*Route{}
	err := PostgresUtils.PrepareQuery(ctx, SELECT_ROUTES_BY_PAGESIZE, &res, Count{Count: (page - 1) * size, Size: size})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (d *DevOpsDbImpl) GetServicesByPageSize(ctx context.Context, page int, size int) ([]*Service, error) {
	res := []*Service{}
	err := PostgresUtils.PrepareQuery(ctx, SELECT_SERVICES_BY_PAGESIZE, &res, Count{Count: (page - 1) * size, Size: size})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (d *DevOpsDbImpl) GetProjectsByPageSize(ctx context.Context, page int, size int) ([]*Project, error) {
	res := []*Project{}
	err := PostgresUtils.PrepareQuery(ctx, SELECT_PROJECTS_BY_PAGESIZE, &res, Count{Count: (page - 1) * size, Size: size})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (d *DevOpsDbImpl) GetProjectsByFuzzyFind(ctx context.Context, fuzzystr string) ([]*Project, error) {
	res := []*Project{}
	err := PostgresUtils.PrepareQuery(ctx, SELECT_PROJECT_BY_FUZZY, &res, FuzzyData{Fuzzystr: fuzzystr})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (d *DevOpsDbImpl) GetDeploymentByFuzzyFind(ctx context.Context, fuzzystr string, namespaceId int, projectId int) ([]*Deployment, error) {
	res := []*Deployment{}
	querySql := SELECT_DEPLOYMENT_BY_FUZZY_PRO_AND_NAMESPACE
	if namespaceId == 0 {
		querySql = strings.ReplaceAll(querySql, "and namespace_id=:namespace_id", "")
	}
	err := PostgresUtils.PrepareQuery(ctx, querySql, &res, struct {
		Fuzzystr    string `db:"fuzzystr"`
		NamespaceId int    `db:"namespace_id"`
		ProjectId   int    `db:"project_id"`
	}{Fuzzystr: fuzzystr, ProjectId: projectId, NamespaceId: namespaceId})
	return res, err
}

func (d *DevOpsDbImpl) GetAllService(ctx context.Context) ([]*Service, error) {
	srvs := []*Service{}
	err := PostgresUtils.PrepareQuery(ctx, "select * from service where 1 = 1", &srvs, Service{})
	if err != nil {
		return nil, err
	}
	return srvs, nil
}

func (d *DevOpsDbImpl) GetAllDeployments(ctx context.Context) ([]*Deployment, error) {
	res := []*Deployment{}
	err := PostgresUtils.PrepareQuery(ctx, "select * from deployment where 1 = 1", &res, Deployment{})
	if err != nil {
		return nil, err
	}
	return res, nil
}
func (d *DevOpsDbImpl) GetAllPro(ctx context.Context) ([]*Project, error) {
	res := []*Project{}
	err := PostgresUtils.PrepareQuery(ctx, "select * from projects where 1=1", &res, Project{})
	if err != nil {
		return nil, err
	}
	return res, nil
}
func (d *DevOpsDbImpl) GetDeploymentsByProAndChannel(ctx context.Context, projectId int, channel string) ([]*Deployment, error) {
	deployments := []*Deployment{}
	err := PostgresUtils.PrepareQuery(ctx, SELECT_DEPLOYMENTS_BY_PRO_AND_CHANNEL, &deployments, Deployment{ProjectId: projectId, ChannelName: channel})
	if err != nil {
		logrus.Errorf(err.Error())
		return nil, err
	}
	return deployments, nil
}
func (d *DevOpsDbImpl) GetRoutesByProAndChannel(ctx context.Context, projectId int, channel string) ([]*Route, error) {
	routes := []*Route{}
	err := PostgresUtils.PrepareQuery(ctx, SELECT_ROUTES_BY_PRO_AND_CHANNEL, &routes, Route{ProjectId: projectId, Channel: channel})
	if err != nil {
		return nil, err
	}
	return routes, nil
}
func (d *DevOpsDbImpl) GetClusterById(ctx context.Context, id int) (*K8sClusterInfo, error) {
	k8sClusterInfo := &K8sClusterInfo{}
	err := PostgresUtils.PrepareQueryRow(ctx, SELECT_CLUSTER_BY_ID, k8sClusterInfo, K8sClusterInfo{Id: id})
	if err != nil {
		return nil, err
	}
	return k8sClusterInfo, nil
}
func (d *DevOpsDbImpl) GetNspByClusterAndName(ctx context.Context, clusterId int, name string) (*K8sNamespace, error) {
	namespace := &K8sNamespace{}
	err := PostgresUtils.PrepareQueryRow(ctx, SELECT_NSP_BY_CLUSTER_ID_AND_NAME, namespace, K8sNamespace{K8sClusterId: clusterId, Name: name})
	if err != nil {
		return nil, err
	}
	return namespace, nil
}

func (d *DevOpsDbImpl) GetProIdByName(ctx context.Context, name string) (*Project, error) {
	res := &Project{}
	err := PostgresUtils.PrepareQueryRow(ctx, SELECT_PROJECT_ID_BY_NAME, res, Project{ProjectName: name})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (d *DevOpsDbImpl) GetProjectByName(ctx context.Context, name string) (*Project, error) {
	res := &Project{}
	err := PostgresUtils.PrepareQueryRow(ctx, SELECT_PROJECT_BY_NAME, res, Project{ProjectName: name})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (d *DevOpsDbImpl) GetClusterIdByName(ctx context.Context, name string) (*K8sClusterInfo, error) {
	res := &K8sClusterInfo{}
	err := PostgresUtils.PrepareQueryRow(ctx, SELECT_CLUSTER_ID_BY_NAME, res, K8sClusterInfo{Name: name})
	if err != nil {
		return nil, err
	}
	return res, nil
}
func (d *DevOpsDbImpl) GetServicesByDeploymentIdAndChannel(ctx context.Context, deploymentId int, channel string) (*Service, error) {
	res := new(Service)
	err := PostgresUtils.PrepareQueryRow(ctx, SELECT_SERVICES_BY_DEPLOYMENT_ID_AND_CHANNEL, res, Service{DeploymentId: deploymentId, ChannelName: channel})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (d *DevOpsDbImpl) GetRoutesByProId(ctx context.Context, projectId int) ([]*Route, error) {
	res := []*Route{}
	err := PostgresUtils.PrepareQuery(ctx, SELECT_ROUTES_BY_PRO_ID, &res, Route{ProjectId: projectId})
	if err != nil {
		return nil, err
	}
	return res, nil
}
func (d *DevOpsDbImpl) GetRoutesRepNamespaceIdByProId(ctx context.Context, projectId int) ([]*Route, error) {
	res := []*Route{}
	err := PostgresUtils.PrepareQuery(ctx, SELECT_ROUTES_REP_NAMESPACE_BY_PROJECT_ID, &res, Route{ProjectId: projectId})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (d *DevOpsDbImpl) GetEnabledBranchsAndTokenByProName(ctx context.Context, projectName string) (*Project, error) {
	res := &Project{}
	err := PostgresUtils.PrepareQueryRow(ctx, SELECT_PROJECT_ENABLED_BRANCHS_AND_TOKEN_BY_PROJECT_NAME, res, Project{ProjectName: projectName})
	if err != nil {
		return nil, err
	}
	return res, nil
}
func (d *DevOpsDbImpl) GetRouteIdsByCondition(ctx context.Context, condition string) {

}

// INSERT_INTO_DEPLOYMENT    = "insert into deployment (deployment_name,project_id,channel_name,content,enabled,docker_repo_id) values(?,?,?,?,?,?)"
// INSERT_INTO_DOCKER_INFO   = "insert into docker_info (name,type,username,password,registry_url,namespace) values(?,?,?,?,?,?)"
// INSERT_INTO_K8S_CLUSTER   = "insert into k8s_cluster (name,ca,token,url,ip) values(?,?,?,?,?)"

// INSERT_INTO_ROUTES        = "insert into routes (project_id,ref_rep,channel,cluster_id,nsp_name,docker,create_time,update_time)"
func (d *DevOpsDbImpl) InsertIntoDeployment(ctx context.Context, deployment Deployment) (*Deployment, error) {
	res := &Deployment{}
	err := PostgresUtils.PrepareQueryRow(ctx, INSERT_INTO_DEPLOYMENT, res, deployment)
	if err != nil {
		return nil, err
	}
	return res, nil
}
func (d *DevOpsDbImpl) InsertIntoDockerInfo(ctx context.Context, dockerInfo DockerInfo) (*DockerInfo, error) {
	res := &DockerInfo{}
	err := PostgresUtils.PrepareQueryRow(ctx, INSERT_INTO_DOCKER_INFO, res, dockerInfo)
	if err != nil {
		return nil, err
	}
	return res, nil
}


func (d *DevOpsDbImpl) InsertIntoK8sCluster(ctx context.Context, k8sClusterInfo K8sClusterInfo) (*K8sClusterInfo, error) {
	res := &K8sClusterInfo{}
	err := PostgresUtils.PrepareQueryRow(ctx, INSERT_INTO_K8S_CLUSTER, res, k8sClusterInfo)
	if err != nil {
		return nil, err
	}
	return res, nil
}
func (d *DevOpsDbImpl) InsertIntoRoutes(ctx context.Context, route Route) (*Route, error) {
	res := &Route{}
	err := PostgresUtils.PrepareQueryRow(ctx, INSERT_INTO_ROUTES, res, route)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (d *DevOpsDbImpl) InsertIntoOpsHistory(ctx context.Context, opsHistory OpsHistory) (*OpsHistory, error) {
	res := &OpsHistory{}
	err := PostgresUtils.PrepareQueryRow(ctx, INSERT_INTO_OPS_HISTORY, res, opsHistory)
	return res, err
}
func (d *DevOpsDbImpl) GetOpsHistoryByPageSize(ctx context.Context, count int, size int) ([]*OpsHistory, error) {
	res := []*OpsHistory{}
	err := PostgresUtils.PrepareQuery(ctx, SELECT_OPS_HISTORY_BY_PAGESIZE, &res, Count{})
	return res, err
}

// INSERT_INTO_K8S_NAMESPACE = "insert into k8s_namespace (k8s_cluster_id,name,description,config) values(?,?,?,?)"
func (d *DevOpsDbImpl) InsertIntoK8sNamespace(ctx context.Context, nsp K8sNamespace) (*K8sNamespace, error) {
	res := &K8sNamespace{}
	err := PostgresUtils.PrepareQueryRow(ctx, INSERT_INTO_K8S_NAMESPACE, res, nsp)
	if err != nil {
		return nil, err
	}
	return res, nil
}

//INSERT_INTO_PROJECTS      = "insert into projects (repo_url,project_name,repo_type,tags,enabled,project_token)"
func (d *DevOpsDbImpl) InsertIntoProjects(ctx context.Context, project Project) (*Project, error) {
	res := &Project{}
	err := PostgresUtils.PrepareQueryRow(ctx, INSERT_INTO_PROJECTS, res, project)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (d *DevOpsDbImpl) InsertIntoService(ctx context.Context, service Service) (*Service, error) {
	res := &Service{}
	err := PostgresUtils.PrepareQueryRow(ctx, INSERT_INTO_SERVICE, res, service)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (d *DevOpsDbImpl) InsertIntoConfig(ctx context.Context, config Config) (*Config, error) {
	res := &Config{}
	err := PostgresUtils.PrepareQueryRow(ctx, INSERT_INTO_CONFIG, res, config)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (d *DevOpsDbImpl) InsertIntoRequestHistory(ctx context.Context, requestHistory RequestHistory) (*RequestHistory, error) {
	res := &RequestHistory{}
	err := PostgresUtils.PrepareQueryRow(ctx, INSERT_INTO_REQUEST_HISTORY, res, requestHistory)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (d *DevOpsDbImpl) GetReqHistoryByPageSize(ctx context.Context, count int, size int) ([]*RequestHistory, error) {
	res := []*RequestHistory{}
	err := PostgresUtils.PrepareQuery(ctx, SELECT_REQUEST_HISTORY_BY_PAGESIZE, res, Count{Count: count, Size: size})
	return res, err
}

func (d *DevOpsDbImpl) GetRouteByProIdAndChannelAndRepAndNspAndCluster(
	ctx context.Context,
	projectId int,
	channel string,
	rep string,
	nsp string,
	clusterid int,
) (*Route, error) {
	route := new(Route)
	r := Route{
		ProjectId: projectId,
		RefRep:    rep,
		Channel:   channel,
		ClusterId: clusterid,
		NspName:   nsp,
	}
	err := PostgresUtils.PrepareQueryRow(ctx, SELECT_ROUTE_BY_PRO_AND_CHANNEL_AND_REP_AND_NSP_AND_CLUSTER, route, r)
	if err != nil {
		return nil, err
	}
	return route, nil
}

func (d *DevOpsDbImpl) GetDeploymentByNameAndProAndChannel(ctx context.Context, name string, projectid int, channel string) (*Deployment, error) {
	deployment := new(Deployment)
	de := Deployment{
		DeploymentName: name,
		ProjectId:      projectid,
		ChannelName:    channel,
	}
	err := PostgresUtils.PrepareQueryRow(ctx, SELECT_DEPLYMENT_BY_NAME_AND_PRO_AND_CHANNEL, deployment, de)
	if err != nil {
		return nil, err
	}
	return deployment, nil
}
func (d *DevOpsDbImpl) GetDeploymentByName(ctx context.Context, name string) ([]*Deployment, error) {
	res := []*Deployment{}
	err := PostgresUtils.PrepareQuery(ctx, SELECT_DEPLOYMENTS_BY_NAME, &res, Deployment{DeploymentName: name})
	return res, err
}

func (d *DevOpsDbImpl) GetUserByName(ctx context.Context, name string) (*User, error) {
	user := &User{}
	err := PostgresUtils.PrepareQueryRow(ctx, SELECT_USER_BY_NAME, user, User{Name: name})
	//userarray := []User{}
	//err = PostgresUtils.PrepareQuery(ctx, "select * from users", &userarray, User{})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		} else {
			return nil, err
		}
	}
	return user, nil
}

func (d *DevOpsDbImpl) GetUserByAccount(ctx context.Context, account string) (*User, error) {
	user := &User{}
	err := PostgresUtils.PrepareQueryRow(ctx, SELECT_USER_BY_ACCOUNT, user, User{Account: account})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		} else {
			return nil, err
		}
	}
	return user, nil
}

func (d *DevOpsDbImpl) InsertIntoUser(ctx context.Context, user User) (*User, error) {
	res := &User{}
	err := PostgresUtils.PrepareQueryRow(ctx, INSERT_INTO_USER, res, user)
	if err != nil {
		return nil, err
	}
	return res, nil
}
func (d *DevOpsDbImpl) GetProjectByGitlabId(ctx context.Context, gitlabId int) (*Project, error) {
	res := &Project{}
	err := PostgresUtils.PrepareQueryRow(ctx, SELECT_PROJECT_BY_GITLAB_ID, res, Project{GitlabId: gitlabId})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (d *DevOpsDbImpl) UpdateK8sCluster(ctx context.Context, k8sCluster K8sClusterInfo) (bool, error) {
	res, err := PostgresUtils.PrepareExec(ctx, UPDATE_CLUSTER_INFO, &k8sCluster)
	if err != nil {
		return false, err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	if rows != 1 {
		return false, fmt.Errorf("更新失败！")
	}
	return true, nil
}

func (d *DevOpsDbImpl) GetNamespaceById(ctx context.Context, id int) (*K8sNamespace, error) {
	res := &K8sNamespace{}
	err := PostgresUtils.PrepareQueryRow(ctx, SELECT_NAMESPACE_BY_ID, res, K8sClusterInfo{Id: id})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (d *DevOpsDbImpl) GetDeploymentByProjectId(ctx context.Context, projectId int) ([]*Deployment, error) {
	res := []*Deployment{}
	err := PostgresUtils.PrepareQuery(ctx, SELECT_DEPLOYMENT_BY_PROJECT_ID, &res, Deployment{ProjectId: projectId})
	return res, err
}

func (d *DevOpsDbImpl) GetConfigByProjectId(ctx context.Context, projectId int, size int, page int) ([]*Config, error) {
	res := []*Config{}
	err := PostgresUtils.PrepareQuery(ctx, SELECT_CONFIGS_BY_PRO_ID, &res, map[string]interface{}{"project_id": projectId, "count": (page - 1) * size, "size": size})
	return res, err
}

func (d *DevOpsDbImpl) GetConfigByProjectIdAndNamespaceId(ctx context.Context, projectId int, namespaceId int, size int, page int) ([]*Config, error) {
	res := []*Config{}
	err := PostgresUtils.PrepareQuery(ctx, SELECT_CONFIGS_BY_PRO_ID_AND_NAMESPACE_ID, &res, map[string]interface{}{"project_id": projectId, "namespace_id": namespaceId, "count": (page - 1) * size, "size": size})
	return res, err
}

func (d *DevOpsDbImpl) GetConfigsCountByProId(ctx context.Context, proid int, nspid int) (int, error) {
	selectSql := SELECT_DEPLOYMENTS_COUNT_BY_PRO_AND_NSP
	if nspid == 0 {
		selectSql = strings.ReplaceAll(SELECT_DEPLOYMENTS_COUNT_BY_PRO_AND_NSP, "and namespace_id=:namespace_id", "")
	}
	c := Count{}
	err := PostgresUtils.PrepareQueryRow(ctx, selectSql, &c, map[string]interface{}{"project_id": proid, "namespace_id": nspid})
	if err != nil {
		return 0, err
	}
	return c.Count, nil
}

func (d *DevOpsDbImpl) GetConfigsCountByProIdAndNspId(ctx context.Context, proid int, nspid int) (int, error) {
	selectSql := SELECT_CONFIGS_COUNT_BY_PRO_AND_NSP
	if nspid == 0 {
		selectSql = strings.ReplaceAll(SELECT_CONFIGS_COUNT_BY_PRO_AND_NSP, "and namespace_id=:namespace_id", "")
	}
	c := Count{}
	err := PostgresUtils.PrepareQueryRow(ctx, selectSql, &c, map[string]interface{}{"project_id": proid, "namespace_id": nspid})
	if err != nil {
		return 0, err
	}
	return c.Count, nil
}

func (d *DevOpsDbImpl) UpdateConfigById(ctx context.Context, config Config) (int, error) {

	res, err := PostgresUtils.PrepareExec(ctx, UPDATE_CONFIG_BY_ID, config)
	if err != nil {
		return 0, err
	}
	affectedRows, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	if affectedRows != 1 {
		return 0, fmt.Errorf("修改未成功！")
	}
	return int(affectedRows), err

}

func (d *DevOpsDbImpl) SelectDingtalkBotByPro(ctx context.Context, projectId int) ([]*DingTalkBot, error) {
	res := []*DingTalkBot{}
	err := PostgresUtils.PrepareQuery(ctx, SELECT_DINGTALK_BOT_BY_PRO, &res, DingTalkBot{ProjectId: projectId})
	return res, err
}

func (d *DevOpsDbImpl) InsertIntoDingtalkBot(ctx context.Context, dingtalkBot DingTalkBot) (*DingTalkBot, error) {
	res := &DingTalkBot{}
	err := PostgresUtils.PrepareQueryRow(ctx, INSERT_INTO_DINGTALK_BOT, res, dingtalkBot)
	return res, err

}

func (d *DevOpsDbImpl) GetConfigByDeploymentId(ctx context.Context, deploymentId int) ([]*Config, error) {
	res := []*Config{}
	err := PostgresUtils.PrepareQuery(ctx, SELECT_CONFIG_BY_DEPLOYMENT_ID, &res, Config{DeploymentId: deploymentId})
	return res, err
}

func (d *DevOpsDbImpl) DeleteConfigById(ctx context.Context, id int) (bool, error) {
	res, err := PostgresUtils.PrepareExec(ctx, DELETE_CONFIG_BY_CONFIG_ID, Config{Id: id})
	if err != nil {
		return false, err
	}
	row, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	if row == 1 {
		return true, nil
	} else if row == 0 {
		return false, errors.New("can't find the config")
	} else {
		return false, errors.New(fmt.Sprintf("delete %d records of config", row))
	}
}

func (d *DevOpsDbImpl) DeleteDingtalkBotById(ctx context.Context, id int) (bool, error) {
	res, err := PostgresUtils.PrepareExec(ctx, DELETE_DINGTALK_BOT, DingTalkBot{Id: id})
	if err != nil {
		return false, err
	}
	row, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	if row == 1 {
		return true, nil
	} else if row == 0 {
		return false, errors.New("can't find the dingtalk_bot")
	} else {
		return false, errors.New(fmt.Sprintf("delete %d records of dingtalk_bot", row))
	}
}

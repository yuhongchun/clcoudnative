package model

import (
	"strings"
	"time"
)

type Project struct {
	Id             int    `json:"id" db:"id"`
	ProjectName    string `json:"project_name" db:"project_name"`
	RepoType       string `json:"repo_type" db:"repo_type"`
	RepoUrl        string `json:"repo_url" db:"repo_url"`
	Tags           string `json:"tags" db:"tags"`
	Enabled        bool   `json:"enabled" db:"enabled"`
	ProjectToken   string `json:"project_token" db:"project_token"`
	EnabledBranchs string `json:"enabled_branchs" db:"enabled_branchs"`
	Descript       string `json:"descript" db:"descript"`
	Topic          string `json:"topic" db:"topic"`
	GitlabId       int    `json:"gitlab_id" db:"gitlab_id"`
	Group          string `json:"group" db:"group"`
}

func (p *Project) ParseTags() map[string]string {
	tags := p.Tags
	tagsSplit := strings.Split(tags, ",")
	res := map[string]string{}
	for _, kv := range tagsSplit {
		kvSplit := strings.Split(kv, "=")
		if len(kvSplit) == 2 {
			res[kvSplit[0]] = kvSplit[1]
		}
	}
	return res
}

type Controller struct {
	Id             int    `json:"id" db:"id"`
	ProjectId      int    `json:"project_id" db:"project_id"`
	DockerMapping  string `json:"docker_mapping" db:"docker_mapping"`
	RefMapping     string `json:"ref_mapping" db:"ref_mapping"`
	ChannelMapping string `json:"channel_mapping" db:"channel_mapping"`
}
type Deployment struct {
	Id             int    `json:"id" db:"id"`
	DeploymentName string `json:"deployment_name" db:"deployment_name"`
	ProjectId      int    `json:"project_id" db:"project_id"`
	ChannelName    string `json:"channel_name" db:"channel_name"`
	Content        string `json:"content" db:"content"`
	Enabled        bool   `json:"enabled" db:"enabled"`
	DockerRepoId   int    `json:"docker_repo_id" db:"docker_repo_id"`
	NamespaceId    int    `json:"namespace_id" db:"namespace_id"`
	NamespaceMsg   string `json:"namespace_msg"`
}
type DockerInfo struct {
	Id          int    `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	Type        string `json:"type" db:"type"`
	Username    string `json:"username" db:"username"`
	Password    string `json:"password" db:"password"`
	RegistryUrl string `json:"registry_url" db:"registry_url"`
	Namespace   string `json:"namespace" db:"namespace"`
}
type K8sClusterInfo struct {
	Id           int    `json:"id" db:"id"`
	Name         string `json:"name" db:"name"`
	Ca           string `json:"ca" db:"ca"`
	Token        string `json:"token" db:"token"`
	EncryptCa    []byte `json:"encrypt_ca" db:"encrypt_ca"`
	EncryptToken []byte `json:"encrypt_token" db:"encrypt_token"`
	Url          string `json:"url" db:"url"`
	Ip           string `json:"ip" db:"ip"`
}

type K8sNamespace struct {
	Id           int    `json:"id" db:"id"`
	K8sClusterId int    `json:"k8s_cluster_id" db:"k8s_cluster_id"`
	Name         string `json:"name" db:"name"`
	Description  string `json:"description" db:"description"`
	Config       string `json:"config" db:"config"`
}

type Service struct {
	Id           int    `json:"id" db:"id"`
	DeploymentId int    `json:"deployment_id" db:"deployment_id"`
	ChannelName  string `json:"channel_name" db:"channel_name"`
	Content      string `json:"content" db:"content"`
}
type Route struct {
	Id          int       `db:"id"`
	ProjectId   int       `db:"project_id"`
	RefRep      string    `db:"ref_rep"`
	Channel     string    `db:"channel"`
	ClusterId   int       `db:"cluster_id"`
	NspName     string    `db:"nsp_name"`
	Docker      string    `db:"docker"`
	Enabled     bool      `db:"enabled"`
	NamespaceId int       `db:"namespace_id"`
	CreateTime  time.Time `db:"create_time"`
	UpdateTime  time.Time `db:"update_time"`
}
type RequestHistory struct {
	Id            int       `json:"id" db:"id"`
	RequestUrl    string    `json:"request_url" db:"request_url"`
	UserId        int       `json:"user_id" db:"user_id"`
	RequestParams string    `json:"request_params" db:"request_params"`
	RequestMethod string    `json:"request_method" db:"request_method"`
	Host          string    `json:"host" db:"host"`
	UpdateTime    time.Time `json:"update_time" db:"update_time"`
}

type OpsHistory struct {
	Id           int       `json:"id" db:"id"`
	ResourceType string    `json:"resource_type" db:"resource_type"`
	OpsType      string    `json:"ops_type" db:"ops_type"`
	ResourceId   int       `json:"resource_id" db:"resource_id"`
	UpdateTime   time.Time `json:"update_time" db:"update_time"`
	ChangeBefore string    `json:"change_before" db:"change_before"`
	ChangeAfter  string    `json:"change_after" db:"change_after"`
	UserId       int       `json:"user_id" db:"user_id"`
}

type Config struct {
	Id              int    `json:"id" db:"id"`
	ProjectId       int    `json:"project_id" db:"project_id"`
	ConfigName      string `json:"config_name" db:"config_name"`
	FileName        string `json:"file_name" db:"file_name"`
	ConfigmapName   string `json:"configmap_name" db:"configmap_name"`
	Content         string `json:"content" db:"content"`
	RestartAfterPub bool   `json:"restart_after_pub" db:"restart_after_pub"`
	NamespaceId     int    `json:"namespace_id" db:"namespace_id"`
	DeploymentId    int    `json:"deployment_id" db:"deployment_id"`
	DeploymentName  string `json:"deployment_name"`
	NamespaceName   string `json:"namespace_name"`
}

type DingTalkBot struct {
	Id              int    `json:"id" db:"id"`
	ProjectId       int    `json:"project_id" db:"project_id"`
	DingTalkBotHook string `json:"dingtalk_bot_hook" db:"dingtalk_bot_hook"`
	Descript        string `json:"descript" db:"descript"`
}

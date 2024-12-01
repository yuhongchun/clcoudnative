package model

import (
	"strings"
	"time"
)

type Project struct {
	Id             int    `mapstructure:"id" db:"id"`
	ProjectName    string `mapstructure:"project_name" db:"project_name"`
	RepoType       string `mapstructure:"repo_type" db:"repo_type"`
	RepoUrl        string `mapstructure:"repo_url" db:"repo_url"`
	Tags           string `mapstructure:"tags" db:"tags"`
	Enabled        bool   `mapstructure:"enabled" db:"enabled"`
	ProjectToken   string `mapstructure:"project_token" db:"project_token"`
	EnabledBranchs string `mapstructure:"enabled_branchs" db:"enabled_branchs"`
	Descript       string `mapstructure:"descript" db:"descript"`
	Topic          string `mapstructure:"topic" db:"topic"`
	GitlabId       int    `db:"gitlab_id"`
	Group          string `db:"group"`
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
	Id             int    `mapstructure:"id" db:"id"`
	ProjectId      int    `mapstructure:"project_id" db:"project_id"`
	DockerMapping  string `mapstructure:"docker_mapping" db:"docker_mapping"`
	RefMapping     string `mapstructure:"ref_mapping" db:"ref_mapping"`
	ChannelMapping string `mapstructure:"channel_mapping" db:"channel_mapping"`
}
type Deployment struct {
	Id             int    `mapstructure:"id" db:"id"`
	DeploymentName string `mapstructure:"deployment_name" db:"deployment_name"`
	ProjectId      int    `mapstructure:"project_id" db:"project_id"`
	ChannelName    string `mapstructure:"channel_name" db:"channel_name"`
	Content        string `mapstructure:"content" db:"content"`
	Enabled        bool   `mapstructure:"enabled" db:"enabled"`
	DockerRepoId   int    `mapstructure:"docker_repo_id" db:"docker_repo_id"`
	NamespaceId    int    `mapstructure:"namespace_id" db:"namespace_id"`
}
type DockerInfo struct {
	Id          int    `mapstructure:"id" db:"id"`
	Name        string `mapstructure:"name" db:"name"`
	Type        string `mapstructure:"type" db:"type"`
	Username    string `mapstructure:"username" db:"username"`
	Password    string `mapstructure:"password" db:"password"`
	RegistryUrl string `mapstructure:"registry_url" db:"registry_url"`
	Namespace   string `mapstructure:"namespace" db:"namespace"`
}
type K8sClusterInfo struct {
	Id           int    `mapstructure:"id" db:"id"`
	Name         string `mapstructure:"name" db:"name"`
	Ca           string `mapstructure:"ca" db:"ca"`
	Token        string `mapstructure:"token" db:"token"`
	Url          string `mapstructure:"url" db:"url"`
	Ip           string `mapstructure:"ip" db:"ip"`
	EncryptCa    []byte `json:"encrypt_ca" db:"encrypt_ca"`
	EncryptToken []byte `json:"encrypt_token" db:"encrypt_token"`
}

type K8sNamespace struct {
	Id           int    `mapstructure:"id" db:"id"`
	K8sClusterId int    `mapstructure:"k8s_cluster_id" db:"k8s_cluster_id"`
	Name         string `mapstructure:"name" db:"name"`
	Description  string `mapstructure:"description" db:"description"`
	Config       string `mapstructure:"config" db:"config"`
}

type Service struct {
	Id           int    `mapstructure:"id" db:"id"`
	DeploymentId int    `mapstructure:"deployment_id" db:"deployment_id"`
	ChannelName  string `mapstructure:"channel_name" db:"channel_name"`
	Content      string `mapstructure:"content" db:"content"`
	Name         string `mapstructure:"name" db:"name"`
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

type Watcher struct {
	RouteId  int    `db:"route_id"`
	UserUUId string `db:"user_uuid"`
}

type Config struct {
	Id              int    `db:"id"`
	ProjectId       int    `db:"project_id"`
	ConfigName      string `db:"config_name"`
	FileName        string `db:"file_name"`
	ConfigmapName   string `db:"configmap_name"`
	Content         string `db:"content"`
	RestartAfterPub bool   `db:"restart_after_pub"`
	NamespaceId     int    `db:"namespace_id"`
	DeploymentId    int    `db:"deployment_id"`
}



type DingTalkBot struct {
	Id              int    `json:"id" db:"id"`
	ProjectId       int    `json:"project_id" db:"project_id"`
	DingTalkBotHook string `json:"dingtalk_bot_hook" db:"dingtalk_bot_hook"`
	Descript        string `json:"descript" db:"descript"`
}

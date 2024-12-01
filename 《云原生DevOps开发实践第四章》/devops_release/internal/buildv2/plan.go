package buildv2

import (
	"context"
	"devops_release/database"
	"devops_release/database/model"
	logr "github.com/sirupsen/logrus"
)

type ClusterInfo struct {
	ClusterName string   `yaml:"cluster_name"`
	Namespaces  []string `yaml:"namespaces"`
}
type ChannelMapping struct {
	Channel string             `yaml:"channel_name"`
	Cluster []K8s_cluster_info `yaml:"clusters"`
}
type ChannelMappingYaml struct {
	ChannelMapping []*ChannelMapping `yaml:"channel_mapping"`
}

// 在router表中确定project_id和deployment_id唯一的资源有没有发版路由
func getReleaseRoutes(ctx context.Context, projectid int, namespaceId int) []*model.Route {
	devopsdb := database.GetDevopsDb()
	routes := []*model.Route{}
	//fmt.Println("namespaceId:")
	//fmt.Println(namespaceId)

	routes, err := devopsdb.GetRoutesByProIdAndNspId(ctx, projectid, namespaceId)
	if err != nil || routes == nil || len(routes) == 0 {
		logr.Errorf("未在系统中找到%d namespaceId:%s的发版路径！请添加！err：", projectid, namespaceId, err)
	}
	return routes
}

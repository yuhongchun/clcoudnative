package project

import (
	"context"
	"fmt"
	"strings"
	"time"

	"devops_build/database"
	"devops_build/database/model"
	"github.com/sirupsen/logrus"
)

type NewProResourceDefault struct {
	ProjectName string `json:"project_name"`
	Type        string `json:"type"`
	Channel     string `json:"channel"`
	ClusterName string `json:"cluster_name"`
	Namespace   string `json:"namespace"`
	Language    string `json:"language"` //选填go|python|java均可
}
type CreateProResourceDefaultRes struct {
	ProjectId    int `json:"project_id"`
	DeploymentId int `json:"deployment_id"`
	ServiceId    int `json:"service_id"`
	RouteId      int `json:"route_id"`
}

func AddNewProResourceFromTemp(ctx context.Context, info NewProResourceDefault) (*CreateProResourceDefaultRes, error) {
	devopsdb := database.GetDevopsDb()
	resInfo := CreateProResourceDefaultRes{}
	projectTemp, err := devopsdb.GetProjectByName(ctx, "project_template")
	if err != nil {
		fmt.Println("no projectTemp")
		return nil, err
	}
	deploymentTemp, err := devopsdb.GetDeploymentByName(ctx, "deployment_template")
	fmt.Println(deploymentTemp[0])
	fmt.Println(deploymentTemp[0].Id,deploymentTemp[0].ChannelName)
	if err != nil || len(deploymentTemp) == 0 {
		fmt.Println("no deploymentTemp")
		return nil, err
	}
	serviceTemp, err := devopsdb.GetServicesByDeploymentIdAndChannel(ctx, deploymentTemp[0].Id, deploymentTemp[0].ChannelName)
	if err != nil {
		fmt.Println("no serviceTemp")
		return nil, err
	}
	routeTemp, err := devopsdb.GetRoutesByProId(ctx, projectTemp.Id)
	if err != nil || len(routeTemp) == 0 {
		fmt.Println("no routeTemp")
		return nil, err
	}

	configmapTemp, err := devopsdb.GetRoutesByProId(ctx, projectTemp.Id)
	if err != nil || len(configmapTemp) == 0 {
		fmt.Println("no configmapTemp")
		return nil, err
	}
	newProject := projectTemp

	newProject.Id = 0
	newProject.ProjectName = info.ProjectName

	proExist, err := devopsdb.IfExistProject(ctx, *newProject)
	var p *model.Project
	if false && (err != nil || proExist) {
		p, err = devopsdb.GetProjectByName(ctx, info.ProjectName)
		if err != nil {
			logrus.Error(err)
			return nil, err
		}
	} else {
		p, err = devopsdb.InsertIntoProjects(ctx, *newProject)
		if err != nil {
			logrus.Error(err)
			return nil, err
		}
	}

	resInfo.ProjectId = p.Id

	newDeployment := deploymentTemp[0]

	newDeployment.Id = 0
	newDeployment.ChannelName = info.Channel
	newDeployment.DeploymentName = info.ProjectName
	newDeployment.ProjectId = p.Id
	newDeployment.Content = strings.ReplaceAll(newDeployment.Content, "${__project_name}", info.ProjectName)
	fmt.Println(newDeployment.Content)
	deploymentExist, err := devopsdb.IfExistDeployment(ctx, *newDeployment)
	if err != nil || deploymentExist {
		return &resInfo, fmt.Errorf("资源已经存在！value:%v", newDeployment)
	}
	d, err := devopsdb.InsertIntoDeployment(ctx, *newDeployment)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	resInfo.DeploymentId = d.Id

	newService := serviceTemp
	newService.ChannelName = info.Channel
	newService.Id = 0
	newService.DeploymentId = d.Id
	newService.Content = strings.ReplaceAll(newService.Content, "${__project_name}", info.ProjectName)
	serviceExist, err := devopsdb.IfExistService(ctx, *newService)
	if err != nil || serviceExist {
		return &resInfo, fmt.Errorf("资源已经存在！value:%v", newService)
	}
	s, err := devopsdb.InsertIntoService(ctx, *newService)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	resInfo.ServiceId = s.Id

	newRoute := routeTemp[0]
	newRoute.Id = 0
	newRoute.Channel = info.Channel

	clusterInfo, err := devopsdb.GetClusterIdByName(ctx, info.ClusterName)
	if err != nil {
		logrus.Error(err)
		return &resInfo, err
	}
	newRoute.ClusterId = clusterInfo.Id
	// 默认允许main分支是可以发版，后续再根据实际情况整改
	newRoute.RefRep = "main"
	newRoute.NspName = info.Namespace
	newRoute.ProjectId = p.Id
	newRoute.CreateTime = time.Now()
	newRoute.UpdateTime = time.Now()
	routeExist, err := devopsdb.IfExistRoute(ctx, *newRoute)
	if err != nil || routeExist {
		return &resInfo, fmt.Errorf("资源已经存在！value:%v", newService)
	}
	r, err := devopsdb.InsertIntoRoutes(ctx, *newRoute)
	if err != nil {
		logrus.Error(err)
		return &resInfo, err
	}
	resInfo.RouteId = r.Id


	return &resInfo, nil
}
package k8sresource

import (
	"context"
	"fmt"

	"devops_release/database"
	"devops_release/database/model"
	nlog "github.com/sirupsen/logrus"
	apiappsv1 "k8s.io/api/apps/v1"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

func SyncDeploymentToLocal(ctx context.Context, deploymentId int, d apiappsv1.Deployment) error {
	if len(d.Name) == 0 {
		return fmt.Errorf("k8sdeployment未部署！")
	}
	d.Status = appsv1.DeploymentStatus{}
	d.Annotations = nil
	d.ManagedFields = nil
	d.CreationTimestamp = v1.Time{}
	d.Namespace = ""
	d.ResourceVersion = ""
	d.SelfLink = ""
	d.UID = ""
	d.Spec.Template.Annotations = nil
	b, err := yaml.Marshal(d)
	if err != nil {
		nlog.Error("同步失败！")
		return err
	}
	fmt.Println("b的值为:",string(b))
	deploymentYaml := `apiVersion: apps/v1
kind: Deployment
` + string(b)
	devopsdb := database.GetDevopsDb()
	// 会从Kubernetes集群中拉取最新deployment信息，修改更新时间，去除多余的字段，然后同步到本地的数据库中
	_, err = devopsdb.UpdateDeploymentContentById(ctx, model.Deployment{
		Id:      deploymentId,
		Content: deploymentYaml,
	})
	return err
}

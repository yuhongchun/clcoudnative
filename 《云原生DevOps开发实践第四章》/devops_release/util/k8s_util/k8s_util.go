package k8sutil

import (
	"context"

	"devops_release/database"
	aesutil "devops_release/util/aes_util"
	apiappsv1 "k8s.io/api/apps/v1"
	appsv1 "k8s.io/client-go/applyconfigurations/apps/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func GetK8sClientByName(ctx context.Context, clusterName string) (*kubernetes.Clientset, error) {
	rest.InClusterConfig()
	devopsdb := database.GetDevopsDb()
	clusterId, err := devopsdb.GetClusterIdByName(context.Background(), clusterName)
	if err != nil {
		return nil, err
	}
	return GetK8sClientById(ctx, clusterId.Id)
}
func GetK8sClientById(ctx context.Context, clusterId int) (*kubernetes.Clientset, error) {
	devopsdb := database.GetDevopsDb()
	clusterInfo, err := devopsdb.GetClusterById(context.Background(), clusterId)
	if err != nil {
		return nil, err
	}
	realToken := clusterInfo.Token
	realCa := clusterInfo.Ca
	if len(realToken) == 0 || len(realCa) == 0 {
		realToken, err = aesutil.DecryptString(clusterInfo.EncryptToken)
		realCa, err = aesutil.DecryptString(clusterInfo.EncryptCa)
		if err != nil {
			return nil, err
		}
	}
	clientset, err := kubernetes.NewForConfig(&rest.Config{
		Host: clusterInfo.Url,
		//Host:            "https://10.251.22.10:6443",
		TLSClientConfig: rest.TLSClientConfig{CAData: []byte(realCa), ServerName: "kubernetes"},
		BearerToken:     realToken,
	})
	if err != nil {
		return nil, err
	}
	return clientset, nil
}

func GetContainerIndexByName(name string, deployment interface{}) int {
	if d, ok := deployment.(appsv1.DeploymentApplyConfiguration); ok {
		for i, c := range d.Spec.Template.Spec.Containers {
			if *c.Name == name {
				return i
			}
		}
	} else if d, ok := deployment.(apiappsv1.Deployment); ok {
		for i, c := range d.Spec.Template.Spec.Containers {
			if c.Name == name {
				return i
			}
		}
	}
	return 0
}

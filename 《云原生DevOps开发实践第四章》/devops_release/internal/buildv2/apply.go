package buildv2

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"devops_release/database/model"
	k8sresource "devops_release/internal/service/k8s_resource"
	k8sutil "devops_release/util/k8s_util"
	"github.com/pkg/errors"
	logr "github.com/sirupsen/logrus"
	apiappsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"
	appsv1 "k8s.io/client-go/applyconfigurations/apps/v1"
	corev1 "k8s.io/client-go/applyconfigurations/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
)

func (spec *UpdateSpec) applyConfigmap(ctx context.Context, configmapString string, clientset *kubernetes.Clientset) error {
	if len(configmapString) != 0 {
		ctx, ctxCancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer ctxCancel()
		configmap := corev1.ConfigMapApplyConfiguration{}
		d := yaml.NewYAMLToJSONDecoder(bytes.NewBufferString(configmapString))

		err := d.Decode(&configmap)
		if err != nil {
			logr.WithContext(ctx).Error(err)
			return err
		}
		// configmap.WithNamespace(spec.Namespace).WithName(spec.ProjectName).WithAPIVersion("v1")
		_, err = clientset.CoreV1().ConfigMaps(spec.K8sNamespace.Name).Apply(ctx, &configmap, v1.ApplyOptions{
			Force:        true,
			FieldManager: "nighting-release"})
		if err != nil {
			logr.WithContext(ctx).Error(err)
			return err
		}
	} else {
		logr.WithContext(ctx).Infof("skip configmap:%s\n", spec.Project.ProjectName)
	}
	return nil
}
func (spec *UpdateSpec) applyDeployment(ctx context.Context, deploymentString string, clientset *kubernetes.Clientset) error {
	if len(deploymentString) != 0 {
		deploymentYaml := appsv1.DeploymentApplyConfiguration{}
		d := yaml.NewYAMLToJSONDecoder(bytes.NewBufferString(deploymentString))
		e := d.Decode(&deploymentYaml)
		if e != nil {
			logr.Error("解析yaml有问题，请注意")
			logr.WithContext(ctx).Error(e)
			return nil
		}
		ctx, ctxCancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer ctxCancel()
		containerIndex := k8sutil.GetContainerIndexByName(*deploymentYaml.Name, deploymentYaml)
		img := strings.Split(*deploymentYaml.Spec.Template.Spec.Containers[containerIndex].Image, ":")
		if len(img) < 2 {
			logr.WithContext(ctx).Infof("skip deployment:%s\n", spec.Project.ProjectName)
			return errors.New("Error Deployment:wrong image format")
		}
		// 获取kubectl apply操作的镜像名:版本号,这里Docker仓库要注意域名或IP的区别
		//image := img[0] + ":" + spec.Version
		image := img[0] + ":" + img[1] + ":" + spec.Version

		if len(spec.Version) != 0 {
			deploymentYaml.Spec.Template.Spec.Containers[containerIndex].Image = &image
		}
		//如果nighting-build从job log中读到了imag信息，则用此镜像信息
		if strings.Contains(spec.Version, ":") {
			image = spec.Version
			deploymentYaml.Spec.Template.Spec.Containers[containerIndex].Image = &spec.Version
		}

		logr.Infof("Applying image:%s to namespace: [%s]", *deploymentYaml.Spec.Template.Spec.Containers[containerIndex].Image, spec.K8sNamespace.Name)

		_, e = clientset.AppsV1().Deployments(spec.K8sNamespace.Name).Apply(ctx, &deploymentYaml, v1.ApplyOptions{Force: true, FieldManager: "nighting-release"})

		if e != nil {
			logr.WithContext(ctx).Infof("skip deployment:%s\n", spec.Project.ProjectName)
			return e
		}
	} else {
		logr.WithContext(ctx).Infof("skip deployment:%s\n", spec.Project.ProjectName)
	}
	return nil
}
func (spec *UpdateSpec) patchDeployment(ctx context.Context, deployment *apiappsv1.Deployment, localDeployment model.Deployment, clientset *kubernetes.Clientset) error {
	k8sContaniers := deployment.Spec.Template.Spec.Containers
	if len(k8sContaniers) == 0 {
		return fmt.Errorf("deployment中没有镜像信息！")
	}
	imageUrl := k8sContaniers[0].Image
	imageUrlSplit := strings.Split(imageUrl, ":")
	if len(imageUrlSplit) < 2 {
		logr.WithContext(ctx).Infof("skip deployment:%s\n", spec.Project.ProjectName)
		fmt.Println(imageUrl)
		return errors.New("Error Deployment:wrong image format")
	}
	// 这里Docker仓库要注意域名或IP的区别
	imageUrl = imageUrlSplit[0] + ":" + imageUrlSplit[1] +":" + spec.Version
	//如果是devops-build应用从job log中读到了imag信息，则用此镜像信息
	if strings.Contains(spec.Version, ":") {
		imageUrl = spec.Version
	}
	containerIndex := k8sutil.GetContainerIndexByName(deployment.Name, *deployment)
	deployment.Spec.Template.Spec.Containers[containerIndex].Image = imageUrl
	if len(k8sContaniers) == 0 {
		return errors.New("k8s containers is empty!")
	}
	//初始化image pathch 将其镜像设置为更新后的值
	imagePatch := []k8sresource.ImageJsonPatch{
		{
			Op:    "replace",
			Path:  fmt.Sprintf("/spec/template/spec/containers/%d/image", containerIndex),
			Value: imageUrl,
		},
	}
	imagePatchBytes, err := json.Marshal(imagePatch)
	if err != nil {
		logr.Error("patch marshal json failed!err:", err)
		return fmt.Errorf("patch marshal json failed!err:%v", err)
	}
	logr.Infof("Patching image:%s to namespace: [%s]", imageUrl, spec.K8sNamespace.Name)
	_, err = clientset.AppsV1().Deployments(spec.K8sNamespace.Name).Patch(ctx, deployment.Name, types.JSONPatchType, imagePatchBytes, v1.PatchOptions{FieldManager: "nighting-release"})
	if err != nil {
		logr.Error("patch image failed!err:", err)
		return fmt.Errorf("patch image failed!err:%v", err)
	}
	k8sresource.SyncDeploymentToLocal(ctx, localDeployment.Id, *deployment)
	return nil
}

func (spec *UpdateSpec) applyService(ctx context.Context, serviceString string, clientset *kubernetes.Clientset) error {
	if len(serviceString) != 0 {
		serviceYaml := corev1.ServiceApplyConfiguration{}
		d := yaml.NewYAMLToJSONDecoder(bytes.NewBufferString(serviceString))
		e := d.Decode(&serviceYaml)
		if e != nil {
			logr.WithContext(ctx).Infof("skip service:%s\n", spec.Project.ProjectName)

			return e
		}
		ctx, ctxCancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer ctxCancel()
		_, e = clientset.CoreV1().Services(spec.K8sNamespace.Name).Apply(ctx, &serviceYaml, v1.ApplyOptions{FieldManager: "nighting-release"})
		if e != nil {
			logr.WithContext(ctx).Infof("skip service:%s\n", spec.Project.ProjectName)
			return e
		}
	} else {
		logr.WithContext(ctx).Infof("skip service:%s\n", spec.Project.ProjectName)
	}
	return nil
}
func (s *UpdateSpec) restartDeployment(ctx context.Context, deploymentString string, clientset *kubernetes.Clientset) error {
	if len(deploymentString) != 0 {
		deploymentYaml := appsv1.DeploymentApplyConfiguration{}
		d := yaml.NewYAMLToJSONDecoder(bytes.NewBufferString(deploymentString))
		e := d.Decode(&deploymentYaml)
		if e != nil {
			logr.WithContext(ctx).Error(e)

			return nil
		}
		ctx, ctxCancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer ctxCancel()

		// Start to Restart!
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			if deploymentYaml.Name == nil {
				return errors.Errorf("键值对不合法！")
			}
			result, getErr := clientset.AppsV1().Deployments(s.K8sNamespace.Name).Get(context.TODO(), *deploymentYaml.Name, v1.GetOptions{})
			if getErr != nil {
				logr.WithContext(ctx).Errorf("Failed to get latest version of Deployment: %v", getErr)
				return getErr
			}
			// 设置重启Deployment输出日志注解
			t := time.Now().Format("2006-01-02 15:04:05")
			m := make(map[string]string)
			m["devops.com/restartedAt"] = t
			result.Spec.Template.Annotations = m
			containerIndex := k8sutil.GetContainerIndexByName(result.Name, *result)
			image := result.Spec.Template.Spec.Containers[containerIndex].Image
			logr.WithContext(ctx).Infof(
				"Restart Deployment %s with image:%s in namespace: [%s], because of the config has changed",
				*deploymentYaml.Name,
				image,
				s.K8sNamespace.Name)
			_, updateErr := clientset.AppsV1().Deployments(s.K8sNamespace.Name).Update(context.TODO(), result, v1.UpdateOptions{FieldManager: "nighting-release"})
			return updateErr
		})
		if retryErr != nil {
			logr.WithContext(ctx).Errorf("Update failed: %v", retryErr)
			return retryErr
		}
		logr.WithContext(ctx).Infof("Restarted deployment %s in cluster[%s]-namespace[%s] success", s.Project.ProjectName, s.K8sClusterName, s.Project.ProjectName)
	} else {
		logr.WithContext(ctx).Infof("skip deployment:%s\n", s.Project.ProjectName)
	}
	return nil
}

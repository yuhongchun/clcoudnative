package compare

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"devops_release/database"
	"devops_release/util/apollo"
	k8sutil "devops_release/util/k8s_util"
	"devops_release/util/model"
	logr "github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type CompareConfigRes struct {
	Config1 map[string]string `json:"apollo_config"`
	Config2 map[string]string `json:"k8s_config"`
	IsEqual bool              `json:"is_equal"`
}

func CompareApolloConfigAndK8s(ctx context.Context, configId int) (*CompareConfigRes, error) {
	devopsdb := database.GetDevopsDb()
	apolloConfigMapping, err := devopsdb.GetConfigById(ctx, configId)
	if err != nil || apolloConfigMapping == nil {
		logr.Error(err)
		return nil, fmt.Errorf("未找到apollo_config_mapping!err:", err)
	}
	fmt.Println(apolloConfigMapping)
	nspInfo, err := devopsdb.GetNamespaceById(ctx, apolloConfigMapping.NamespaceId)
	if err != nil {
		logr.Error(err)
		return nil, err
	}
	clusterInfo, err := devopsdb.GetClusterById(ctx, nspInfo.K8sClusterId)
	if err != nil {
		logr.Error(err)
		return nil, err
	}
	apolloItems, err := apollo.GetRealKVSFromApollo(ctx, apolloConfigMapping.ProjectId, apolloConfigMapping.ConfigName, clusterInfo.Name, nspInfo.Name)
	if err != nil {
		logr.Error(err)
		logr.Error("获取apollo配置失败！err:", err)
	}

	clientset, err := k8sutil.GetK8sClientById(ctx, nspInfo.K8sClusterId)
	if err != nil {
		logr.Error(err)
		return nil, fmt.Errorf("获取k8sclient出错！err:%v", err)
	}

	k8s_items := getConfigFromK8s(clientset, nspInfo.Name, apolloConfigMapping.ConfigName)
	if k8s_items == nil {
		logr.Error("获取k8s配置失败！")
	}
	if apolloItems == nil {
		apolloItems = []model.Item{}
	}
	if k8s_items == nil {
		k8s_items = []model.Item{}
	}
	compareRes := compareConfig(apolloItems, k8s_items)
	return compareRes, nil
}

func compareConfig(config1 []model.Item, config2 []model.Item) *CompareConfigRes {
	compareConfigRes := &CompareConfigRes{
		Config1: map[string]string{},
		Config2: map[string]string{},
	}

	config1Map := itemsToMap(config1)
	config2Map := itemsToMap(config2)
	selectUnique(config1, config2Map, compareConfigRes.Config1)
	selectUnique(config2, config1Map, compareConfigRes.Config2)
	selectDiff(config1Map, config2Map, compareConfigRes.Config1, compareConfigRes.Config2)
	if len(compareConfigRes.Config1) == 0 && len(compareConfigRes.Config2) == 0 {
		compareConfigRes.IsEqual = true
	}
	return compareConfigRes
}
func itemsToMap(items []model.Item) map[string]string {
	m := map[string]string{}
	for _, item := range items {
		m[item.Key] = item.Value
	}
	return m
}
func selectUnique(config []model.Item, configMap map[string]string, compareRes map[string]string) {
	for _, item := range config {
		if _, ok := configMap[item.Key]; !ok {
			compareRes[item.Key] = item.Value
		}
	}

}
func selectDiff(config1 map[string]string, config2 map[string]string, res1 map[string]string, res2 map[string]string) {
	for k, v := range config1 {
		if config1[k] != config2[k] {
			res1[k] = v
			res2[k] = config2[k]
		}
	}
}

func getConfigFromK8s(clientset *kubernetes.Clientset, namespace string, cmName string) []model.Item {
	res, err := clientset.CoreV1().ConfigMaps(namespace).Get(context.Background(), cmName, v1.GetOptions{})
	if err != nil {
		fmt.Println(err)
		return nil
	}
	settings := res.Data["settings.yaml"]
	return ConvertYamlToKv(settings)
}

func ConvertYamlToKv(content string) []model.Item {
	kvs, err := apollo.YamlTransitionApollo(content)
	if err != nil {
		fmt.Println("yaml 非法")
		fmt.Println(content)
		return nil
	}
	items := []model.Item{}
	for k, v := range kvs {
		value := ""
		switch v.(type) {
		case int:
			value = strconv.Itoa(v.(int))
		case string:
			value = v.(string)
		case bool:
			value = strconv.FormatBool(v.(bool))
		default:
			value = ""
		}
		items = append(items, model.Item{
			Key:   strings.TrimSpace(k),
			Value: value,
		})
	}
	return items
}

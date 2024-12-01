package apollo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"devops_release/database"
	"devops_release/util/model"
	"devops_release/util/myyaml"
	logr "github.com/sirupsen/logrus"

	sysconfig "devops_release/config"
)

type OpenApi struct {
	Env            string
	AddressOpenapi string
	Appid          string
	Token          string
	ClusterName    string
	NamespaceName  string
}


func (o *OpenApi) GetAllClusterInfo() ([]model.EnvMap, error) {
	url := fmt.Sprintf("%s/openapi/v1/apps/%s/envclusters", o.AddressOpenapi, o.Appid)
	headers := map[string]string{
		"Authorization": o.Token,
		"Content-Type":  "application/json;charset=UTF-8",
	}
	data, err := sendRequest("GET", headers, url, nil)
	if err != nil {
		logr.Errorf("请求失败！err:", err)
		return nil, err
	}
	envMaps := []model.EnvMap{}
	err = json.Unmarshal(data, &envMaps)
	if err != nil {
		logr.Errorf("获取clusterInfo失败！url:%s", url)
		logr.Errorf("json解析失败！json:%s err:%v", string(data), err)
		return nil, err
	}
	fmt.Println(envMaps)
	return envMaps, nil
}

// http://{portal_address}/openapi/v1/envs/{env}/apps/{appId}/clusters/{clusterName}/namespaces
func (o *OpenApi) GetAllNamespaceInfos() ([]model.Namespace, error) {
	url := fmt.Sprintf("%s/openapi/v1/envs/%s/apps/%s/clusters/%s/namespaces", o.AddressOpenapi, o.Env, o.Appid, o.ClusterName)
	headers := map[string]string{
		"Authorization": o.Token,
		"Content-Type":  "application/json;charset=UTF-8",
	}
	data, err := sendRequest("GET", headers, url, nil)
	if err != nil {
		logr.Errorf("请求失败！err:", err)
		return nil, err
	}
	namespaces := []model.Namespace{}
	err = json.Unmarshal(data, &namespaces)
	if err != nil {
		logr.Errorf("json解析失败！content:%s err:%v", string(data), err)
		return nil, err
	}
	//fmt.Println(namespaces)
	return namespaces, nil
}

// http://{portal_address}/openapi/v1/envs/{env}/apps/{appId}/clusters/{clusterName}/namespaces/{namespaceName}
// 获取apollo项目配置  此处namespace对应projectName
func (o *OpenApi) GetNamespaceInfo(namespaceName string) (*model.Namespace, error) {
	url := fmt.Sprintf("%s/openapi/v1/envs/%s/apps/%s/clusters/%s/namespaces/%s", o.AddressOpenapi, o.Env, o.Appid, o.ClusterName, namespaceName)
	fmt.Println(url)
	headers := map[string]string{
		"Authorization": o.Token,
		"Content-Type":  "application/json;charset=UTF-8",
	}
	data, err := sendRequest("GET", headers, url, nil)
	if err != nil {
		logr.Errorf("请求失败！err:", err)
		return nil, err
	}
	namepace := model.Namespace{}
	err = json.Unmarshal(data, &namepace)
	if err != nil {
		logr.Error("json解析失败！err:", err)
		return nil, err
	}
	return &namepace, nil
}

func (o *OpenApi) AddApolloCluster(name string) error {
	url := fmt.Sprintf("%s/openapi/v1/envs/%s/apps/%s/clusters", o.AddressOpenapi, o.Env, o.Appid)
	cluster := Cluster{
		Name:                name,
		AppId:               o.Appid,
		DataChangeCreatedBy: "yuhongchun",
	}
	headers := map[string]string{
		"Authorization": o.Token,
		"Content-Type":  "application/json;charset=UTF-8",
	}
	clusterBytes, err := json.Marshal(cluster)
	if err != nil {
		return err
	}
	date, err := sendRequest("POST", headers, url, clusterBytes)
	if err != nil {
		return err
	}
	logr.Infof("add apollo cluster info:", date)
	return nil
}
func (o *OpenApi) AddNameSpace(ctx context.Context, name string) error {
	url := fmt.Sprintf("%s/openapi/v1/apps/%s/appnamespaces", o.AddressOpenapi, o.Appid)
	namespaceInfo := NamespaceInfo{
		Name:                name,
		AppId:               o.Appid,
		Format:              "properties",
		IsPublic:            false,
		DataChangeCreatedBy: "yuhongchun",
	}
	headers := map[string]string{
		"Authorization": o.Token,
		"Content-Type":  "application/json;charset=UTF-8",
	}
	namespaceBytes, err := json.Marshal(namespaceInfo)
	if err != nil {
		return err
	}
	data, err := sendRequest("POST", headers, url, namespaceBytes)
	if err != nil {
		return err
	}
	logr.Info("add apollo namespace info:", data)
	return nil
}
func (o *OpenApi) AddItem(item model.Item) error {
	url := fmt.Sprintf("%s/openapi/v1/envs/%s/apps/%s/clusters/%s/namespaces/%s/items", o.AddressOpenapi, o.Env, o.Appid, o.ClusterName, o.NamespaceName)
	itemsInfo := Config{
		Key:                 item.Key,
		Value:               item.Value,
		DataChangeCreatedBy: "nighting-release",
	}
	headers := map[string]string{
		"Authorization": o.Token,
		"Content-Type":  "application/json;charset=UTF-8",
	}
	itemsInfoBytes, err := json.Marshal(itemsInfo)
	if err != nil {
		return err
	}
	data, err := sendRequest("POST", headers, url, itemsInfoBytes)
	if err != nil {
		return err
	}
	logr.Info("add apollo item info:", string(data))

	logr.Info("token", o.Token)
	return nil
}
func (o *OpenApi) AddItems(items []model.Item) []error {
	errs := []error{}
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	wg.Add(len(items))
	for _, item := range items {
		go func(i model.Item) {
			err := o.AddItem(i)
			if err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			}
			wg.Done()
		}(item)
	}
	return errs
}

func (o *OpenApi) UpdateItem(item model.Item) error {
	url := fmt.Sprintf("%s/openapi/v1/envs/%s/apps/%s/clusters/%s/namespaces/%s/items/%s?createIfNotExists=true", o.AddressOpenapi, o.Env, o.Appid, o.ClusterName, o.NamespaceName, item.Key)
	itemsInfo := UpdateInfo{
		Key:                      item.Key,
		Value:                    item.Value,
		DataChangeLastModifiedBy: "nighting-build",
		DataChangeCreatedBy:      "nighting-build",
	}
	headers := map[string]string{
		"Authorization": o.Token,
		"Content-Type":  "application/json;charset=UTF-8",
	}
	itemsInfoBytes, err := json.Marshal(itemsInfo)
	if err != nil {
		return err
	}
	data, err := sendRequest("PUT", headers, url, itemsInfoBytes)
	if err != nil {
		return err
	}
	logr.Info("update apollo item info:", string(data))

	logr.Info("token", o.Token)
	return nil
}

func (o *OpenApi) UpdateItems(items []model.Item) []error {
	errs := []error{}
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	wg.Add(len(items))
	for _, item := range items {
		go func(i model.Item) {
			err := o.UpdateItem(i)
			if err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			}
			wg.Done()
		}(item)
	}
	return errs
}

func (o *OpenApi) ReleaseConfig() error {
	url := fmt.Sprintf("%s/openapi/v1/envs/%s/apps/%s/clusters/%s/namespaces/%s/releases", o.AddressOpenapi, o.Env, o.Appid, o.ClusterName, o.NamespaceName)
	config := Release{
		ReleaseTitle:   time.Now().Format("2006-01-02 15:04:05"),
		ReleaseComment: "",
		ReleasedBy:     "nighting-build",
	}
	configBytes, err := json.Marshal(config)
	if err != nil {
		logr.Error("json解析错误！err:", err)
		return err
	}
	headers := map[string]string{
		"Authorization": o.Token,
		"Content-Type":  "application/json;charset=UTF-8",
	}
	res, err := sendRequest("POST", headers, url, configBytes)
	if err != nil {
		logr.Error("发布apollo配置失败！err:", err)
		return err
	}
	logr.Info("release info:", res)
	return nil
}

func GetYamlFromApollo(ctx context.Context, projectId int, configName string, cluster string, namespace string) (string, error) {
	kvs, err := GetKVSFromApollo(ctx, projectId, configName, cluster, namespace)
	if err != nil {
		return "", err
	}
	myyaml := myyaml.NewYaml(kvs)
	content := myyaml.ToString()
	return content, nil
}

func GetKVSFromApollo(ctx context.Context, projectId int, configName string, cluster string, namespace string) ([]model.Item, error) {
	ch := GetApChannelFromConfigByClusterName(cluster)
	if ch == nil {
		logr.Error("apollo没有配置该集群！")
		return nil, fmt.Errorf("apollo没有配置该集群！")
	}
	devopsdb := database.GetDevopsDb()

	pro, err := devopsdb.GetProjectById(ctx, projectId)
	if err != nil {
		return nil, err
	}
	var apolloApp sysconfig.App
	for _, app := range ch.Apps {
		if app.Type == pro.ParseTags()["type"] {
			apolloApp = app
			break
		}
	}

	op := OpenApi{
		Env:            ch.Key,
		AddressOpenapi: ch.AddressOpenapi,
		Appid:          apolloApp.Id,
		ClusterName:    namespace,
		Token:          apolloApp.Token,
		NamespaceName:  configName,
	}

	msp, err := op.GetNamespaceInfo(op.NamespaceName)
	if err != nil {
		return nil, err
	}
	return msp.Items, nil
}

func GetRealKVSFromApollo(ctx context.Context, projectId int, configName string, cluster string, namespace string) ([]model.Item, error) {
	ch := GetApChannelFromConfigByClusterName(cluster)
	if ch == nil {
		logr.Error("apollo没有配置该集群！")
		return nil, fmt.Errorf("apollo没有配置该集群！")
	}
	devopsdb := database.GetDevopsDb()

	pro, err := devopsdb.GetProjectById(ctx, projectId)
	if err != nil {
		return nil, err
	}
	var apolloApp sysconfig.App
	for _, app := range ch.Apps {
		if app.Type == pro.ParseTags()["type"] {
			apolloApp = app
			break
		}
	}

	op := OpenApi{
		Env:            ch.Key,
		AddressOpenapi: ch.AddressOpenapi,
		Appid:          apolloApp.Id,
		ClusterName:    namespace,
		Token:          apolloApp.Token,
		NamespaceName:  configName,
	}

	msp, err := op.GetNamespaceInfo(op.NamespaceName)
	if err != nil {
		return nil, err
	}
	logr.Info("获取公共配置！")
	commonNspInfo, err := op.GetNamespaceInfo(CommonNsp)
	if err != nil {
		logr.Error("获取common配置失败！skip... err：", err)
		commonNspInfo = &model.Namespace{}
	}

	for i, item := range msp.Items {
		//用公共配置填充变量
		realValue := item.Value
		for _, commonItem := range commonNspInfo.Items {
			realValue = strings.ReplaceAll(realValue, fmt.Sprintf("${%s}", commonItem.Key), commonItem.Value)
		}
		msp.Items[i].Value = realValue
		item.Key = strings.TrimSpace(item.Key)
	}
	fmt.Println("------------------------------------")
	fmt.Println(msp.Items)

	return msp.Items, nil
}

func sendRequest(method string, headers map[string]string, url string, body []byte) ([]byte, error) {
	ioreader := bytes.NewBuffer(body)
	ctx, cancle := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancle()
	req, err := http.NewRequestWithContext(ctx, method, url, ioreader)
	if err != nil {
		logr.Errorf("创建请求错误！err:%v", err)
		return nil, err
	}
	for k, v := range headers {
		req.Header.Add(k, v)
	}
	cli := http.Client{}
	resp, err := cli.Do(req)
	if err != nil {
		logr.Errorf("发送请求失败！url:%s err:", url, err)
		return nil, err
	}
	respBody := resp.Body
	defer respBody.Close()
	data, err := ioutil.ReadAll(respBody)
	if err != nil {
		logr.Error("读取响应体错误！err:", err)
		return nil, err
	}
	return data, nil
}

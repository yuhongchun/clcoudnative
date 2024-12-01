package apollo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/apolloconfig/agollo/v4"
	"github.com/apolloconfig/agollo/v4/env/config"
	"github.com/apolloconfig/agollo/v4/storage"
	cfg "devops_release/config"
	"devops_release/util/model"
	logr "github.com/sirupsen/logrus"
)

func writeConfig(namespace string, client *agollo.Client) {
	cache := client.GetConfigCache(namespace)
	cache.Range(func(key, value interface{}) bool {
		fmt.Println("key : ", key, ", value :", value)
		return true
	})
}

type CustomChangeListener struct {
	Env        string
	Cluster    string
	CallBack   RestartCallBack
	Appid      string
	K8sCluster string
	ApolloKey  string
}
type RestartCallBack func(ctx context.Context, k8sCluster string, nsp_name string, projectName string) error

//监听apollo接口的实现
func (c *CustomChangeListener) OnChange(changeEvent *storage.ChangeEvent) {

	//TODO: 校验 通知
	logr.Infof("监听到apollo配置中心配置变更！appid:%s env:%s cluster:%s namespace:%s pro:%s", c.Appid, c.Env, c.K8sCluster, c.Cluster, changeEvent.Namespace)
	if changeEvent.Namespace == CommonNsp {
		logr.Infof("监听到apollo公共配置变更！env:%s clusterName:%s", c.Env, c.Cluster)
		openApi := OpenApi{
			Appid:       c.Appid,
			Env:         c.Env,
			ClusterName: c.Cluster,
		}
		channel := GetApChannelFromConfigByClusterName(c.K8sCluster)
		if channel == nil {
			logr.Errorf("%s不在配置中！", c.Env)
			return
		}
		if channel.Key == strings.ToLower(c.Env) {
			for _, a := range channel.Apps {
				if a.Id == c.Appid {
					openApi.Token = a.Token
					openApi.AddressOpenapi = channel.AddressOpenapi
				}
			}

		}

		namespaceInfos, err := openApi.GetAllNamespaceInfos()
		if err != nil {
			logr.Errorf("公共配置变更，获取namesapceInfos失败！err:", err)
			return
		}
		//检查是否引用了该公共配置
		for _, n := range namespaceInfos {
			for _, item := range n.Items {
				for k, _ := range changeEvent.Changes {
					if strings.Contains(item.Value, k) {
						logr.Info("引用了公共变量，重启...")
						err = c.CallBack(context.Background(), c.K8sCluster, c.Cluster, changeEvent.Namespace)
						if err != nil {
							logr.Errorf("重启服务失败!namespace:%s err:%v", n.NamespaceName, err)
						}
					}
				}
			}
		}
	} else {
		err := c.CallBack(context.Background(), c.K8sCluster, c.Cluster, changeEvent.Namespace)
		if err != nil {
			logr.Errorf("重启服务失败!%v", err)
		}
	}

}

func (c *CustomChangeListener) OnNewestChange(event *storage.FullChangeEvent) {
	//write your code here
	// for key, value := range event.Changes {
	// 	fmt.Println("new key : ", key, ", value :", value)
	// }
}
func StartListening(callBack RestartCallBack) {
	//对不同环境的不同app进行监听
	for _, channel := range cfg.ApolloConfig.Channel {
		logr.Infof("对%s通道进行监听！address_openapi:%s address_sdk:%s ", channel.Key, channel.AddressOpenapi, channel.AddressSDK)
		for _, app := range channel.Apps {
			logr.Infof("%s通道内监听!appid:%s", channel.Key, app.Id)
			appOps := model.AppOps{
				AppId:          app.Id,
				AddressOpenapi: channel.AddressOpenapi,
				AddressSDK:     channel.AddressSDK,
				Token:          app.Token,
				Secret:         app.Secret,
				Env:            []string{channel.Key},
			}
			startListening(channel.Key, channel.K8sCluster, appOps, callBack)
		}
	}

}
func startListening(apolloKey string, k8sClusterName string, appops model.AppOps, callBack RestartCallBack) {
	openApi := OpenApi{
		Appid:          appops.AppId,
		AddressOpenapi: appops.AddressOpenapi,
		Token:          appops.Token,
	}
	envmaps, err := openApi.GetAllClusterInfo()
	if err != nil {
		logr.Errorf("获取环境集群信息失败!err:%v", err)
		panic(err)
	}
	//此循环按照当前设计只有1次
	for _, envmap := range envmaps {
		openApi.Env = envmap.Env
		//对apollo中的不同集群进行监听
		for _, cluster := range envmap.Clusters {
			logr.Infof("对%s集群进行监听！", cluster)
			openApi.ClusterName = cluster
			namespaces, err := openApi.GetAllNamespaceInfos()
			if err != nil {
				logr.Errorf("获取%s环境%s集群下的namespaces失败！err:%v", openApi.Env, cluster, err)
				continue
			}
			namespaceNames := ""
			//拼接namespace
			for i, namespace := range namespaces {
				namespaceNames = namespaceNames + namespace.NamespaceName
				if i == len(namespaces)-1 {
					continue
				}
				namespaceNames = namespaceNames + ","
			}
			//监听这个集群下所有的namespace（project）
			go listeningOfCluster(appops, apolloKey, cluster, k8sClusterName, namespaceNames, false, callBack)

		}
	}
}
func listeningOfCluster(app model.AppOps, apolloKey string, clusterName string, k8sClusterName string, namespaces string, isBackUp bool, callBack RestartCallBack) {
	o := OpenApi{
		Appid:          app.AppId,
		ClusterName:    clusterName,
		AddressOpenapi: app.AddressOpenapi,
		Token:          app.Token,
		Env:            app.Env[0],
	}
	client, err := GetClient(app.AppId, clusterName, app.AddressSDK, namespaces, false, app.Secret)
	if err != nil {
		logr.Errorf("获取apollo_client错误!appid:%s clusterName:%s address:%s namespaceNames:%s err:%v", app.AppId, clusterName, app.AddressSDK, namespaces, err)
	}
	logr.Infof("监听成功！clustername:%s", clusterName)
	//监听变更
	listener := &CustomChangeListener{}
	listener.Env = app.Env[0]
	listener.Cluster = clusterName
	listener.CallBack = callBack
	listener.Appid = app.AppId
	listener.K8sCluster = k8sClusterName
	listener.ApolloKey = apolloKey
	client.AddChangeListener(listener)

	//十秒检查一次apollo，看有没有新的项目添加，有新的就加入监听。
	t := time.NewTicker(10 * time.Second)
	for true {
		select {
		case <-t.C:
			// nlog.Infof("更新监听的namespaces！")
			namespacesNew, err := o.GetAllNamespaceInfos()
			if err != nil {
				// nlog.Errorf("获取新的namespaceinfos失败！appid:%s clusterName:%s err：%v", appId, clusterName, err)
				continue
			}
			newNamespaceNames := ""
			for i, newN := range namespacesNew {
				newNamespaceNames = newNamespaceNames + newN.NamespaceName
				if i == len(namespacesNew)-1 {
					continue
				}
				newNamespaceNames = newNamespaceNames + ","
			}
			newNamespaces := strings.Split(newNamespaceNames, ",")
			for _, newNamespace := range newNamespaces {
				if !strings.Contains(namespaces, newNamespace) {
					logr.Info("发现新增namespace：", newNamespace)
					newCli, err := GetClient(app.AppId, clusterName, app.AddressSDK, newNamespace, false, app.Secret)
					if err != nil {
						logr.Errorf("监听新的namespace失败！newnamespace:", newNamespace)
					}
					namespaces = namespaces + "," + newNamespace
					newListener := &CustomChangeListener{
						Env:        app.Env[0],
						Cluster:    clusterName,
						CallBack:   callBack,
						Appid:      app.AppId,
						K8sCluster: k8sClusterName,
						ApolloKey:  apolloKey,
					}
					newCli.AddChangeListener(newListener)
				}
			}

		}
	}
}

func GetClient(appid string, cluster string, ip string, namepaceNames string, isBackupConfig bool, secret string) (*agollo.Client, error) {
	cfg := &config.AppConfig{
		AppID:          appid,
		Cluster:        cluster,
		IP:             ip,
		NamespaceName:  namepaceNames,
		IsBackupConfig: isBackupConfig,
		Secret:         secret,
	}
	client, err := agollo.StartWithConfig(func() (*config.AppConfig, error) {
		return cfg, nil
	})
	if err != nil {
		logr.Errorf("获取%s/%s/%s的client失败!err:%v", appid, cluster, namepaceNames, err)
		return nil, err
	}

	return client, nil
}

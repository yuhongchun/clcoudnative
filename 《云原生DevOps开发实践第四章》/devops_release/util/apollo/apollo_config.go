package apollo

import (
	"context"
	"fmt"
	"os"
	"strings"

	"devops_release/config"
	"devops_release/database"
	"devops_release/util/model"
	"devops_release/util/myyaml"
	"github.com/sirupsen/logrus"
)

const (
	ConfigmapHeader = `
apiVersion: v1
kind: ConfigMap
metadata:
  name: ${projectName}
data:
  setting.yaml: |-`
	CommonNsp = "common"
)


func GetConfigFromApolloV2(env string, cluster string, projectname string, class string) (string, error) {
	var apolloEnv string
	var apolloChannel config.Channel
	var apolloApp config.App
	var nsp_name string
	nsp_name = projectname
	//var projectName dbmodel.Project
	//var nsp_name = projectNameConfig.ProjectName
	//var nsp_name = apolloConfig.ConfigName
	for _, apchannel := range config.ApolloConfig.Channel {

		if env == apchannel.Key {
			apolloEnv = env
			apolloChannel = apchannel
			break
		}

	}
	if len(apolloEnv) == 0 {
		return "", fmt.Errorf("%s未在配置中！", env)
	}
	for _, app := range apolloChannel.Apps {
		if app.Type == class {
			apolloApp = app
			break
		}
	}
	if len(apolloApp.Id) == 0 {
		return "", fmt.Errorf("%s未在配置中！", class)
	}
	openApi := OpenApi{
		Env:            strings.ToUpper(apolloEnv),
		Appid:          apolloApp.Id,
		Token:          apolloApp.Token,
		ClusterName:    cluster,
		AddressOpenapi: apolloChannel.AddressOpenapi,
	}
	logrus.Infof("从apollo拉取配置|appid:%s cluster:%s name:%s", apolloApp.Id, cluster, nsp_name)
	namespaceInfo, err := openApi.GetNamespaceInfo(nsp_name)
	if err != nil {
		logrus.Errorf("未在配置中心找到%s的配置！跳过...  err:%v", nsp_name, err)
		return "", nil
	}
	items := namespaceInfo.Items
	if items == nil || len(namespaceInfo.Items) == 0 {
		return "", nil
	}
	myaml := myyaml.NewYaml(items)
	yamlString := myaml.ToString()
	if len(yamlString) == 0 {
		logrus.Errorf("yaml格式错误！")
		return "", fmt.Errorf("yaml格式错误！")
	}
	yamlString = "\n" + yamlString

	//configmapHeader := strings.ReplaceAll(ConfigmapHeader, "${projectName}", apolloConfig.ConfigmapName)
	configmapHeader := strings.ReplaceAll(ConfigmapHeader, "${projectName}", nsp_name)
	yamlString = strings.ReplaceAll(yamlString, "\n", "\n    ")
	configmap := configmapHeader + yamlString
	logrus.Infof("从apollo拉取公共配置!appid:%s cluster:%s name:%s", apolloApp.Id, cluster, nsp_name)
	commonNspInfo, err := openApi.GetNamespaceInfo(CommonNsp)
	if err != nil {
		logrus.Error("获取common配置失败！skip... err：", err)
		commonNspInfo = &model.Namespace{}
	}
	for _, item := range commonNspInfo.Items {
		configmap = strings.ReplaceAll(configmap, fmt.Sprintf("${%s}", item.Key), item.Value)
	}
	configmap = strings.ReplaceAll(configmap, "${__project_name}", nsp_name)
	configmap = strings.ReplaceAll(configmap, "${__k8s_namespace}", cluster)
	//configmap = strings.ReplaceAll(configmap, "${fileName}", apolloConfig.FileName)
	//configmap = strings.ReplaceAll(configmap, ":${fileName}", nsp_name)
	configmap = strings.Trim(configmap, myyaml.Space)
	configmap = strings.Trim(configmap, myyaml.Newline)
	fmt.Println("*********************configmap****************************")
	fmt.Println(configmap)
	fmt.Println("***********************end********************************")
	testFileName := "view.yaml"
	file, err := os.Create(testFileName)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()
	file.WriteString(configmap)


	//取此configMap对应的关系，写进config表中，这段逻辑需要重新设计
	/*
	devopsdb := database.GetDevopsDb()
	ctx := context.Background()

	fmt.Println(nsp_name,cluster)
	project, err := devopsdb.GetProjectByName(ctx, nsp_name)

	if err != nil {
		panic(err)
	}
	k8scluster,err := devopsdb.GetNspByClusterAndName(ctx,1,cluster)
	if err != nil {
		panic(err)
	}
	deployment, err := devopsdb.GetDeploymentByName(ctx,nsp_name)
	if err != nil {
		panic(err)
	}

	fmt.Println(project.Id,k8scluster.Id,deployment[0].Id)
	res := dbmodel.Config{}
	res.DeploymentId = deployment[0].Id
	res.ProjectId = project.Id
	res.NamespaceId = k8scluster.Id
	res.Content = configmap
	res.RestartAfterPub = true
	// configMap在apollo中的取名，这里建议取唯一命名
	res.FileName = project.ProjectName
	res.ConfigName = project.ProjectName
	res.ConfigmapName = project.ProjectName

	fmt.Println("当前新增的Deployment名字为：",deployment[0].Id)
	result,err := devopsdb.GetConfigByDeploymentId(ctx,deployment[0].Id)
	fmt.Println(result)
	if len(result) == 0 {
		_, err = devopsdb.InsertIntoConfig(ctx, res)
		if err != nil {
			panic(err)
		}
	}
	*/

	return configmap, nil
}

func getAppByType(c config.Channel, type1 string) (config.App, error) {
	for _, a := range c.Apps {
		if a.Type == type1 {
			return a, nil
		}
	}
	return config.App{}, fmt.Errorf("%s类型不在配置中，检查项目tag的type属性！", type1)
}

//从配置获取发版通道对应环境的apollo
func GetApChannelFromConfigByClusterName(clusterName string) *config.Channel {
	for _, apchannel := range config.ApolloConfig.Channel {
		if clusterName == apchannel.K8sCluster {
			return &apchannel
		}
	}
	return nil
}

func PubApolloConfig(ctx context.Context, configId int) error {
	logrus.Info("release apollo config")
	devopsdb := database.GetDevopsDb()
	apolloConfigMapping, err := devopsdb.GetConfigById(ctx, configId)
	if err != nil {
		return err
	}
	nsp, err := devopsdb.GetNamespaceById(ctx, apolloConfigMapping.NamespaceId)
	if err != nil {
		return err
	}
	cluster, err := devopsdb.GetClusterById(ctx, nsp.K8sClusterId)
	if err != nil {
		return err
	}
	project, err := devopsdb.GetProjectById(ctx, apolloConfigMapping.ProjectId)
	if err != nil {
		return err
	}
	op := OpenApi{}
	apChannel := GetApChannelFromConfigByClusterName(cluster.Name)
	apApp, err := getAppByType(*apChannel, project.ParseTags()["type"])
	if err != nil {
		return err
	}
	op.AddressOpenapi = apChannel.AddressOpenapi
	op.Appid = apApp.Id
	op.Token = apApp.Token
	op.Env = apChannel.Key
	op.NamespaceName = apolloConfigMapping.ConfigName
	op.ClusterName = nsp.Name
	err = op.ReleaseConfig()
	return err
}

func SetApConfig(ctx context.Context, configId int, content string, fileType string) error {
	devopsdb := database.GetDevopsDb()
	apolloConfigMapping, err := devopsdb.GetConfigById(ctx, configId)
	if err != nil {
		return err
	}
	nsp, err := devopsdb.GetNamespaceById(ctx, apolloConfigMapping.NamespaceId)
	if err != nil {
		return err
	}
	cluster, err := devopsdb.GetClusterById(ctx, nsp.K8sClusterId)
	if err != nil {
		return err
	}
	project, err := devopsdb.GetProjectById(ctx, apolloConfigMapping.ProjectId)
	if err != nil {
		return err
	}
	op := OpenApi{}
	apChannel := GetApChannelFromConfigByClusterName(cluster.Name)
	apApp, err := getAppByType(*apChannel, project.ParseTags()["type"])
	if err != nil {
		return err
	}
	op.AddressOpenapi = apChannel.AddressOpenapi
	op.Appid = apApp.Id
	op.Token = apApp.Token
	op.Env = apChannel.Key
	op.NamespaceName = apolloConfigMapping.ConfigName
	op.ClusterName = nsp.Name

	apConfigInfo, err := op.GetNamespaceInfo(project.ProjectName)
	if err != nil || apConfigInfo == nil {
		logrus.Info("未在apollo创建该配置,", project.ProjectName)
		err = op.AddNameSpace(ctx, project.ProjectName)
		if err != nil {
			logrus.Error("创建apollo namespace错误!")
			return err
		}
	}

	curItems := []model.Item{}

	if fileType == "yaml" {
		kvs, err := YamlTransitionApollo(content)
		if err != nil {
			logrus.Error(err)
			return err
		}
		for k, v := range kvs {
			item := model.Item{Key: k}
			switch v.(type) {
			case string:
				item.Value = v.(string)
				isnumber := IsNumber(v.(string))
				if isnumber {
					item.Value = `"` + v.(string) + `"`
				} else {
					item.Value = SetContainerSpecialCh(v.(string))
				}

			case int:
				item.Value = fmt.Sprintf("%d", v.(int))
			case int64:
				item.Value = fmt.Sprintf("%d", v.(int64))
			case float32:
				item.Value = fmt.Sprintf("%f", v.(float32))
			case float64:
				fmt.Println("float64")
				item.Value = fmt.Sprintf("%f", v.(float64))
			case bool:
				item.Value = fmt.Sprintf("%t", v.(bool))
			}

			curItems = append(curItems, item)
		}
	} else if fileType == "kv" {
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			lineSplit := strings.Split(line, "=")
			key := ""
			value := ""
			if len(lineSplit) != 2 {
				key = strings.TrimSpace(lineSplit[0])
			} else {
				key = strings.TrimSpace(lineSplit[0])
				value = strings.TrimSpace(lineSplit[1])
			}
			item := model.Item{
				Key:   key,
				Value: value,
			}
			curItems = append(curItems, item)
		}
	}
	errs := op.UpdateItems(curItems)
	if len(errs) != 0 {
		errMsg := ""
		for _, e := range errs {
			errMsg = errMsg + e.Error() + "\n"
		}
		return fmt.Errorf(errMsg)
	}
	return nil
}

func IsNumber(s string) bool {
	for _, r := range s {
		if 48 <= r && r <= 57 {

		} else {
			return false
		}
	}
	return true
}

func SetContainerSpecialCh(s string) string {
	special := map[rune]bool{
		':': true, '{': true, '}': true, '[': true, ']': true, ',': true, '&': true, '*': true, '#': true, '?': true, '|': true, '-': true, '<': true, '>': true, '=': true, '!': true, '%': true, '"': true, '\'': true,
	}
	for _, r := range s {
		if r == '"' {
			s = "'" + s + "'"
			return s
		}
		if r == '\'' {
			s = `"` + s + `"`
			return s
		}
	}
	for i, r := range s {
		if i == 0 {
			if _, ok := special[r]; ok {
				s = `"` + s + `"`
				return s
			}
		}
	}
	return s
}

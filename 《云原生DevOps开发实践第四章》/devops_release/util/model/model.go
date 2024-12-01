package model

type EnvMap struct {
	Env      string   `mapstructure:"env"`
	Clusters []string `mapstructure:"clusters"`
}
type Namespace struct {
	AppId                      string `mapstructure:"appId"`
	ClusterName                string `mapstructure:"clusterName"`
	NamespaceName              string `mapstructure:"namespaceName"`
	Comment                    string `mapstructure:"comment"`
	Format                     string `mapstructure:"format"`
	IsPublic                   bool   `mapstructure:"isPublic"`
	Items                      []Item `mapstructure:"items"`
	dataChangeCreatedBy        string `mapstructure:"dataChangeCreatedBy"`
	dataChangeLastModifiedBy   string `mapstructure:"dataChangeLastModifiedBy"`
	DataChangeCreatedTime      string `mapstructure:"dataChangeCreatedTime"`
	DataChangeLastModifiedTime string `mapstructure:"dataChangeLastModifiedTime"`
}
type Item struct {
	Key                        string `mapstructure:"key"`
	Value                      string `mapstructure:"value"`
	DataChangeCreatedBy        string `mapstructure:"dataChangeCreatedBy"`
	DataChangeLastModifiedBy   string `mapstructure:"dataChangeLastModifiedBy"`
	DataChangeCreatedTime      string `mapstructure:"dataChangeCreatedTime"`
	DataChangeLastModifiedTime string `mapstructure:"dataChangeLastModifiedTime"`
}

type AppOps struct {
	AppId          string
	AddressOpenapi string
	AddressSDK     string
	Token          string
	Secret         string
	Env            []string
}

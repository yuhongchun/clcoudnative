package config

type ApplicationSettings struct {
	ReadTimeout   int    `mapstructure:"readtimeout"`
	WriterTimeout int    `mapstructure:"writertimeout"`
	Mode          string `mapstructure:"mode"`
	Name          string `mapstructure:"name"`
	Host          string `mapstructure:"host"`
	Port          string `mapstructure:"port"`
	IsHttps       bool   `mapstructure:"ishttps"`
	AesKey        string `mapstructure:"aes_key"`
}

type LoggerSettings struct {
	LogLevel    string `mapstructure:"logLevel"`
	filePath    string `mapstructure:"filePath"`
	MaxFileSize int    `mapstructure:"maxFileSize"`
	MaxBackups  int    `mapstructure:"maxBackups"`
	MaxAge      int    `mapstructure:"maxAge"`
	Compress    bool   `mapstructure:"compress"`
}

type SentrySettings struct {
	Dsn    string `mapstructure:"dsn"`
	Source string `mapstructure:"source"`
}

type EtcdSettings struct {
	Endpoints   []string
	UserName    string
	Password    string
	DialTimeout int
}
type ApolloSettings struct {
	Channel []Channel `mapstructure:"channel"`
}
type Channel struct {
	Key            string `mapstructure:"key"`
	AddressOpenapi string `mapstructure:"address_openapi"`
	AddressSDK     string `mapstructure:"address_sdk"`
	K8sCluster     string `mapstructure:"k8s_cluster"`
	Apps           []App  `mapstructure:"apps"`
}
type App struct {
	Id     string `mapstructure:"id"`
	Token  string `mapstructure:"token"`
	Secret string `mapstructure:"secret"`
	Type   string `mapstructure:"type"`
}
type PostgresSetting struct {
	Dsn string `mapstructure:"dsn"`
}
type NoticerSettings struct {
	Host string `mapstructure:"host"`
}
type TecentCloudSettings struct {
	SecretKey string `mapstructure:"secret_key"`
	SecretId  string `mapstructure:"secret_id"`
}

package config

import (
	"log"

	nlog "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var SentryConfig SentrySettings
var LogConfig LoggerSettings
var ApplicationConfig ApplicationSettings
var EtcdConfig EtcdSettings
var ApolloConfig ApolloSettings
var PostgresConfig PostgresSetting
var NoticerConfig NoticerSettings
var TecentCloudConfig TecentCloudSettings

func Setup(path string) {
	settingCfg := viper.New()
	settingCfg.SetConfigFile(path)
	if err := settingCfg.ReadInConfig(); err != nil {
		log.Fatal("读取配置文件失败.", err)
	}

	// logger
	err := settingCfg.UnmarshalKey("sentry", &SentryConfig)
	if err != nil {
		log.Fatalf("viper.Unmarshal(sentry) faild:%v", err)
	}
	err = settingCfg.UnmarshalKey("apm", &LogConfig)
	if err != nil {
		log.Fatalf("viper.Unmarshal(apm) faild:%v", err)
	}

	//err = logger.Use(&logger.Config{
	//	LogLevel: LogConfig.LogLevel,
	//	APM: &logger.APMConfig{
	//		FilePath:    LogConfig.filePath,
	//		MaxFileSize: LogConfig.MaxFileSize,
	//		MaxBackups:  LogConfig.MaxBackups,
	//		MaxAge:      LogConfig.MaxAge,
	//		Compress:    LogConfig.Compress,
	//	},
	//	Sentry: &logger.SentryConfig{
	//		DSN:    SentryConfig.Dsn,
	//		Source: SentryConfig.Source,
	//	},
	//})
	if err != nil {
		log.Fatalf("logger.Use() faild:%v", err)
	}

	// etcd
	err = settingCfg.UnmarshalKey("etcd", &EtcdConfig)
	if err != nil {
		log.Fatalf("viper.Unmarshal(etcd) faild:%v", err)
	}
	// Server
	err = settingCfg.UnmarshalKey("application", &ApplicationConfig)
	if err != nil {
		log.Fatalf("viper.Unmarshal(application) faild:%v", err)
	}
	nlog.Info(ApplicationConfig)

	err = settingCfg.UnmarshalKey("apollo", &ApolloConfig)
	if err != nil {
		log.Fatalf("viper.Unmarshal(apollo) faild:%v", err)
	}
	err = settingCfg.UnmarshalKey("postgres", &PostgresConfig)
	if err != nil {
		log.Fatalf("viper.Unmarshal(postgres) faild:%v", err)
	}
	err = settingCfg.UnmarshalKey("noticer", &NoticerConfig)
	if err != nil {
		log.Fatalf("viper.Unmarshal(noticer) faild:%v", err)
	}
	err = settingCfg.UnmarshalKey("tecent_cloud", &TecentCloudConfig)
	if err != nil {
		log.Fatalf("viper.Unmarshal(tecent_cloud) faild:%v", err)
	}
}

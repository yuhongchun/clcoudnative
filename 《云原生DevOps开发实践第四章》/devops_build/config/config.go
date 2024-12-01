package config

import (
	"github.com/spf13/viper"
)

func SetUp(path string) {
	settingCfg := viper.New()
	settingCfg.SetConfigFile(path)
	if err := settingCfg.ReadInConfig(); err != nil {
		panic("读取配置文件失败:" + err.Error())
	}

	cfgLogger := settingCfg.Sub("logger")
	if cfgLogger == nil {
		panic("config not found logger")
	}
	LoggerConfig = InitLogger(cfgLogger)

	// 启动参数
	cfgApplication := settingCfg.Sub("application")
	if cfgApplication == nil {
		panic("config not found application")
	}
	ApplicationConfig = InitApplication(cfgApplication)

	cfgRelease := settingCfg.Sub("release")
	if cfgRelease == nil {
		panic("config not found release")
	}
	ReleaseConfig = InitRelease(cfgRelease)

	//gitlab配置初始化
	cfgGitlab := settingCfg.Sub("gitlab")
	if cfgGitlab == nil {
		panic("config not found gitlab")
	}
	GitlabConfig = InitGitlab(cfgGitlab)
	// 钉钉机器人配置文件初始化
	cfgDingtalkRobot := settingCfg.Sub("dingtalk_robot")
	if cfgDingtalkRobot == nil {
		panic("config not found dingtalk_robot")
	}
	DingtalkRobotConfig = InitDingtalkRobot(cfgDingtalkRobot)

	cfgPostgres := settingCfg.Sub("postgres")
	if cfgPostgres == nil {
		panic("config not found cfgpostgres")
	}
	PostgresConfig = InitPostgresConfig(cfgPostgres)
	cfgNightingHost := settingCfg.Sub("nighting_host")
	if cfgNightingHost == nil {
		panic("config not found cfgNightingHost")
	}
	NightingHostConfig = InitNightingHost(cfgNightingHost)
}


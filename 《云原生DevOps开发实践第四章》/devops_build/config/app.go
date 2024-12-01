package config

import "github.com/spf13/viper"

var ApplicationConfig = new(Application)

type Application struct {
	ReadTimeout        int
	WriterTimeout      int
	Mode               string
	Name               string
	Host               string
	Port               string
	IsHttps            bool
	NightingReleaseUrl string
	CorpHashXORKey     string
	AesKey             string
}

func InitApplication(cfg *viper.Viper) *Application {
	return &Application{
		ReadTimeout:        cfg.GetInt("readTimeout"),
		WriterTimeout:      cfg.GetInt("writerTimeout"),
		Host:               cfg.GetString("host"),
		Port:               portDefault(cfg),
		Name:               cfg.GetString("name"),
		Mode:               cfg.GetString("mode"),
		IsHttps:            isHttpsDefault(cfg),
		NightingReleaseUrl: cfg.GetString("nighting_release_url"),
		AesKey:             cfg.GetString("aes_key"),
	}
}

func portDefault(cfg *viper.Viper) string {
	if cfg.GetString("port") == "" {
		return "5000"
	}
	return cfg.GetString("port")
}

func isHttpsDefault(cfg *viper.Viper) bool {
	if cfg.GetString("ishttps") == "" || !cfg.GetBool("ishttps") {
		return false
	}
	return true
}

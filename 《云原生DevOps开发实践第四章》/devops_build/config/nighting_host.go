package config

import "github.com/spf13/viper"

var NightingHostConfig = new(NightingHost)

type NightingHost struct {
	NightingReleaseHost       string `mapstructure:"nighting_release_host"`
	NightingAuthorizationHost string `mapstructure:"nighting_authorization_host"`
	NightingConfigHost        string `mapstructure:"nighting_config_host"`
}

func InitNightingHost(cfg *viper.Viper) *NightingHost {
	return &NightingHost{
		NightingReleaseHost:       cfg.GetString("nighting_release_host"),
		NightingAuthorizationHost: cfg.GetString("nighting_authorization_host"),
		NightingConfigHost:        cfg.GetString("nighting_config_host"),
	}
}

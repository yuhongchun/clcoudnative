package config

import "github.com/spf13/viper"

var ReleaseConfig = new(Release)

type Release struct {
	EnabledChanel []string `json:"enabled_chanel"`
}

func InitRelease(cfg *viper.Viper) *Release {
	return &Release{
		EnabledChanel: cfg.GetStringSlice("enabled_chanel"),
	}
}

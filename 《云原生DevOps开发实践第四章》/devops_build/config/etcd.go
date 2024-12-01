package config

import "github.com/spf13/viper"

var ETCDConfig = new(ETCD)

type ETCD struct {
	Endpoints   []string
	UserName    string
	Password    string
	DialTimeout int
}

func InitETCD(cfg *viper.Viper) *ETCD {
	return &ETCD{
		Endpoints:   cfg.GetStringSlice("endpoints"),
		UserName:    cfg.GetString("username"),
		Password:    cfg.GetString("password"),
		DialTimeout: cfg.GetInt("dialtimeout"),
	}
}

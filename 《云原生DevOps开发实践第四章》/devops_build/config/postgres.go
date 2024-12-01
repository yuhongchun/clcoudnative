package config

import "github.com/spf13/viper"

type PostgresSetting struct {
	Dsn string `mapstructure:"dsn"`
}

var PostgresConfig = new(PostgresSetting)

func InitPostgresConfig(cfg *viper.Viper) *PostgresSetting {
	return &PostgresSetting{Dsn: cfg.GetString("dsn")}
}

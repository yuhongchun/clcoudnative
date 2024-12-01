package config

import "github.com/spf13/viper"

var LoggerConfig = new(Logger)

type Logger struct {
	Sentry   SentryConfig
	APM      APMConfig
	LogLevel string
}

type SentryConfig struct {
	DSN      string
	Source   string
	LogLevel string
}

type APMConfig struct {
	FilePath    string
	MaxFileSize int
	MaxBackups  int
	MaxAge      int
	Compress    bool
}

func InitLogger(cfg *viper.Viper) *Logger {
	return &Logger{
		Sentry: SentryConfig{
			DSN:      cfg.GetString("dsn"),
			Source:   cfg.GetString("source"),
			LogLevel: cfg.GetString("loglevel"),
		},
		APM: APMConfig{
			FilePath:    cfg.Sub("apm").GetString("filepath"),
			MaxFileSize: cfg.Sub("apm").GetInt("maxfilesize"),
			MaxBackups:  cfg.Sub("apm").GetInt("maxbackups"),
			MaxAge:      cfg.Sub("apm").GetInt("maxage"),
			Compress:    cfg.Sub("apm").GetBool("compress"),
		},
		LogLevel: cfg.GetString("loglevel"),
	}
}

package config

import "github.com/spf13/viper"

var GitlabConfig = new(Gitlab)

type Gitlab struct {
	ApiHost           string
	PriToken          string
	CIJobLogDockerKey string
	Keywords          []string
}

func InitGitlab(cfg *viper.Viper) *Gitlab {
	return &Gitlab{
		ApiHost:           cfg.GetString("api_host"),
		PriToken:          cfg.GetString("pri_token"),
		CIJobLogDockerKey: cfg.GetString("ci_job_log_docker_key"),
		Keywords:          cfg.GetStringSlice("keywords"),
	}
}

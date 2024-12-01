package config

import "github.com/spf13/viper"

var DingtalkRobotConfig = new(DingtalkRobot)

type DingtalkRobot struct {
	Token     string            `json:"token"`
	Secret    string            `json:"secret"`
	KeyWords  string            `json:"key_words"`
	MemberMap map[string]string `json:"member_map"`
}

func InitDingtalkRobot(cfg *viper.Viper) *DingtalkRobot {
	return &DingtalkRobot{
		Token:     cfg.GetString("token"),
		Secret:    cfg.GetString("secret"),
		KeyWords:  cfg.GetString("keywords"),
		MemberMap: cfg.GetStringMapString("member_map"),
	}
}

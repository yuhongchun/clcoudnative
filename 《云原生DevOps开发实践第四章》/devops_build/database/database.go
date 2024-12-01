package database

import (
	"github.com/sirupsen/logrus"
	"time"

	"devops_build/config"
	etcd "go.etcd.io/etcd/client/v3"
)

var ETCDClient *etcd.Client

func SetUp() {
	client, err := etcd.New(etcd.Config{
		Endpoints:            config.ETCDConfig.Endpoints,
		Username:             config.ETCDConfig.UserName,
		Password:             config.ETCDConfig.Password,
		DialKeepAliveTime:    5 * time.Second,
		DialKeepAliveTimeout: time.Duration(config.ETCDConfig.DialTimeout) * time.Second,
		DialTimeout:          time.Duration(config.ETCDConfig.DialTimeout) * time.Second,
	})
	if err != nil {
		logrus.Panicf("Panic: ETCD connection failed, error msg: %s ", err)
	}
	// var alive bool
	// var e error
	// for _, endpoint := range config.ETCDConfig.Endpoints {
	// 	_, err := client.Status(context.Background(), endpoint)
	// 	if err == nil {
	// 		alive = true
	// 	} else {
	// 		e = err
	// 	}
	// }
	// if !alive {
	// 	logger.Panic("Panic: ETCD not alive, error msg: %s", e)
	// }
	ETCDClient = client
	logrus.Info("ETCD connection success!")
}

func newClient() (*etcd.Client, error) {
	client, err := etcd.New(etcd.Config{
		Endpoints:            config.ETCDConfig.Endpoints,
		Username:             config.ETCDConfig.UserName,
		Password:             config.ETCDConfig.Password,
		DialKeepAliveTime:    5 * time.Second,
		DialKeepAliveTimeout: time.Duration(config.ETCDConfig.DialTimeout) * time.Second,
		DialTimeout:          time.Duration(config.ETCDConfig.DialTimeout) * time.Second,
	})
	if err != nil {
		logrus.Errorf("Panic: ETCD connection failed, error msg: %s ", err)
		return nil, err
	}
	ETCDClient = client
	logrus.Info("ETCD connection success!")
	return client, nil
}

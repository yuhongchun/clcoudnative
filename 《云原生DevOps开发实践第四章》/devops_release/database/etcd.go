package database

import (
	"time"

	"devops_release/config"
	nlog "github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var etcdClient *clientv3.Client

func GetEtcdClient() *clientv3.Client {
	if etcdClient == nil {
		var err error
		etcdClient, err = clientv3.New(clientv3.Config{
			Endpoints:   []string{config.EtcdConfig.Endpoints[0]},
			Username:    config.EtcdConfig.UserName,
			Password:    config.EtcdConfig.Password,
			DialTimeout: time.Duration(config.EtcdConfig.DialTimeout) * time.Second,
		})
		if err != nil {
			nlog.Panicf("ERROR: ETCD connection failed, error msg: %s ", err)
		}
	}
	return etcdClient
}

package database

// func GetETCDValue(ctx context.Context, key string) (*clientv3.GetResponse, error) {
// 	span, ctx := apm.StartSpan(ctx, "GetETCDValue", "ETCD")
// 	defer span.End()
// 	labels := nlog.LabelFromContext(ctx)

// 	cli, err := newClient()
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer func() {
// 		err := cli.Close()
// 		if err != nil {
// 			nlog.WithContext(ctx).WithLabel(labels).Info("close etcd client failed")
// 		}
// 	}()
// 	nlog.WithContext(ctx).WithLabel(labels).Infof("Info: DialTimeOut: %d", config.ETCDConfig.DialTimeout)
// 	timeOutCtx, cancle := context.WithTimeout(context.Background(), time.Duration(config.ETCDConfig.DialTimeout)*time.Second)
// 	resp, err := cli.Get(timeOutCtx, key)
// 	cancle()
// 	return resp, err
// }

// func WatchETCD() {
//	nlog.Info("start watching etcd changes")
//	rch := ETCDClient.Watch(context.Background(), "/nighting-build/Project", clientv3.WithPrefix())
//	for wresp := range rch {
//		for _, ev := range wresp.Events {
//			nlog.Infof("%s %q \n", ev.Type, ev.Kv.Key)
//			if ev.Type.String() == "PUT" {
//				send.FileChange(context.Background(), string(ev.Kv.Key))
//			}
//		}
//	}
//}

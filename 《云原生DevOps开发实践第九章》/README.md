### 帮助文档

自定义的exporter开发流程为

定义集群指标采集器 --> 数据采集工作 --> 实现describe接口,写入描述信息 --> 将收集的数据导入colletc --> 结构体实例化赋值 --> 定义注册表 --> 注入自定义指标 --> 暴露metrics

exporter_demo.go是一个参考例子，更复杂的跟exporter相关的工程可以参考以下：

[node-exporter](https://github.com/prometheus/node_exporter)
[gpfs-exporter](https://github.com/treydock/gpfs_exporter)
[slurm-exporter]https://github.com/vpenso/prometheus-slurm-exporter

大家可以结合这些开源工程，并结合自己的业务，开发自定义的exporter组件工程。
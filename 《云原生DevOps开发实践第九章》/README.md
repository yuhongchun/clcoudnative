### 帮助文档

自定义的exporter开发流程为

定义集群指标采集器 --> 数据采集工作 --> 实现describe接口,写入描述信息 --> 将收集的数据导入colletc --> 结构体实例化赋值 --> 定义注册表 --> 注入自定义指标 --> 暴露metrics

exporter_demo.go是一个参考例子，更复杂的跟exporter相关的工程可以参考以下：

[node-exporter](https://github.com/prometheus/node_exporter)
[gpfs-exporter](https://github.com/treydock/gpfs_exporter)
[slurm-exporter]https://github.com/vpenso/prometheus-slurm-exporter
[hpc-exporter](https://github.com/SODALITE-EU/hpc-exporter)

大家可以结合这些开源工程，并结合自己的业务，开发自定义的exporter组件工程;Prometheus本身就很强大，如果大家的业务非常重视监控和报警的话，可以深入研究Prometheus原理，像各种recode_rule及Prometheus proxy等等，都可以应用于自己的业务告警，这块甚至都可以单独剥离出来，不仅仅用于云原生监控，也适用于各种虚拟机/物理机/VPN+专线告警。
package main

import (
"github.com/prometheus/client_golang/prometheus"
"github.com/prometheus/client_golang/prometheus/promhttp"
"github.com/siddontang/go/log"
"net/http"
)

//先定义结构体，这是一个集群的指标采集器
type HostMonitor struct {
	cpuDesc    *prometheus.Desc
	memDesc    *prometheus.Desc
	ioDesc     *prometheus.Desc
	labelVaues []string
}

//创建结构体及对应的指标信息
func NewHostMonitor() *HostMonitor {
	return &HostMonitor{
		cpuDesc: prometheus.NewDesc(
			"host_cpu",
			"get host cpu",
			//动态标签key列表
			[]string{"instance_id", "instance_name"},
			//静态标签
			prometheus.Labels{"module": "cpu"},
		),
		memDesc: prometheus.NewDesc(
			"host_mem",
			"get host mem",
			//动态标签key列表
			[]string{"instance_id", "instance_name"},
			//静态标签
			prometheus.Labels{"module": "mem"},
		),
		ioDesc: prometheus.NewDesc(
			"host_io",
			"get host io",
			//动态标签key列表
			[]string{"instance_id", "instance_name"},
			//静态标签
			prometheus.Labels{"module": "io"},
		),
		labelVaues: []string{"myhost", "yunwei"},
	}
}

//实现Describe接口，传递指标描述符到channel
func (h *HostMonitor) Describe(ch chan<- *prometheus.Desc) {
	ch <- h.cpuDesc
	ch <- h.memDesc
	ch <- h.ioDesc
}

//实现collect接口，将执行抓取函数并返回数据
func (h *HostMonitor) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(h.cpuDesc, prometheus.GaugeValue, 70, h.labelVaues...)
	ch <- prometheus.MustNewConstMetric(h.memDesc, prometheus.GaugeValue, 30, h.labelVaues...)
	ch <- prometheus.MustNewConstMetric(h.ioDesc, prometheus.GaugeValue, 90, h.labelVaues...)
}

func main() {


	ovs := NewHostMonitor()
	prometheus.MustRegister(ovs)


	http.Handle("/metrics", promhttp.Handler())

	log.Info("begin to server on port 8080")
	// listen on port 8080
	log.Fatal(http.ListenAndServe(":8081", nil))
}
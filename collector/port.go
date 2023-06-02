package collector

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/host"
)

var (
	isexist   float64 = 1
	namespace         = "own_process"
	endetail          = "datails"
	endmems           = "mems"
)

// 定义收集指标结构体
// 分为进程信息和内存信息
type PortCollector struct {
	ProcessDetail portMetrics
	ProcessMems   portMetrics
	mutex         sync.Mutex // 使用于多个协程访问共享资源的场景
	// value         prometheus.Gauge
}

type portMetrics []struct {
	desc  *prometheus.Desc
	value map[string]string
}

func (p *PortCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range p.ProcessDetail {
		ch <- metric.desc
	}

	for _, metric := range p.ProcessMems {
		ch <- metric.desc
	}
	// ch <- p.ProcessMems
}

func (p *PortCollector) Collect(ch chan<- prometheus.Metric) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	// ch <- prometheus.MustNewConstMetric(p.ProcessMems, prometheus.GaugeValue, 0)
	for _, metric := range p.ProcessDetail {
		ch <- prometheus.MustNewConstMetric(metric.desc, prometheus.GaugeValue, isexist, metric.value["cmdroot"], metric.value["cmdline"], metric.value["Name"], metric.value["State"], metric.value["PPid"], metric.value["Uid"], metric.value["Gid"])
	}
	for _, metric := range p.ProcessMems {
		ch <- prometheus.MustNewConstMetric(metric.desc, prometheus.GaugeValue, isexist, metric.value["Name"], metric.value["pid"], metric.value["VmHWM"], metric.value["VmRSS"])
	}
}

func (p *PortCollector) Updata() {
	// Do nothing here as the value is generated in the Collect() function
}

func newMetrics(p []string, s map[string]map[string]string, u string) *portMetrics {
	host, _ := host.Info()
	hostname := host.Hostname
	var detailList, memsList portMetrics
	for _, v := range p {
		// fmt.Println(k, v)
		detailList = append(detailList, struct {
			desc  *prometheus.Desc
			value map[string]string
		}{
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, v, endetail),
				"Process-related information of port "+v,
				[]string{"cmdroot", "cmdline", "process_name", "status", "ppid", "ownuser", "owngroup"}, // 设置动态labels,collect函数里传来的就是这个变量的值
				prometheus.Labels{"host_name": hostname}),                                               // 设置静态labels
			value: s[v],
		})

		memsList = append(memsList, struct {
			desc  *prometheus.Desc
			value map[string]string
		}{
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, v, endmems),
				"Process memory usage information of port "+v,
				[]string{"process_name", "pid", "vmhwm", "vmrss"}, // 设置动态labels,collect函数里传来的就是这个变量的值
				prometheus.Labels{"host_name": hostname}),         // 设置静态labels
			value: s[v],
		})
	}
	if u == "detail" {
		return &detailList
	} else {
		return &memsList
	}
}

// NewPortCollector 创建port收集器，返回指标信息
func NewPortCollector(p []string, m string) *PortCollector {
	final := GetProcessInfo(p, m)
	// fmt.Printf("test_fanal:%#v", len(final))
	if len(final) == 0 {
		isexist = 0
	} else {
		isexist = 1
	}
	return &PortCollector{
		ProcessDetail: *newMetrics(p, final, "detail"),
		ProcessMems:   *newMetrics(p, final, "mems"),
	}
}

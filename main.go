package main

import (
	"fmt"
	"net/http"
	"time"

	"exporter/collector"

	"github.com/alecthomas/kingpin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// 定义命令行参数
var (
	ticker = kingpin.Flag("ticker", "Interval for obtaining indicators.").Short('t').Default("5").Int()
	mode   = kingpin.Flag("mode", "Using netstat or lsof for specified port pid information.").Short('m').Default("netstat").String()
	port   = kingpin.Flag("port", "This service is to listen the port.").Short('p').Default("9527").String()
	ports  = kingpin.Arg("ports", "The process of listening on ports.").Required().Strings()
)

// func init() {
// 	kingpin.Version("v0.1")
// 	kingpin.Parse()
// 	//注册自身采集器
// 	prometheus.MustRegister(collector.NewPortCollector(*ports, *mode))
// }

func main() {
	kingpin.Version("1.1")
	kingpin.Parse()
	// 注册自身采集器
	prometheus.MustRegister(collector.NewPortCollector(*ports, *mode))
	// fmt.Printf("Would ping: %s with timeout %s \n", *mode, *ports)
	go func() {
		for {
			collector.NewPortCollector(*ports, *mode).Updata()
			time.Sleep(time.Duration(*ticker) * time.Second)
		}
	}()
	http.Handle("/metrics", promhttp.Handler())
	fmt.Println("Ready to listen on port:", *port)
	if err := http.ListenAndServe("0.0.0.0:"+*port, nil); err != nil {
		fmt.Printf("Error occur when start server %v", err)
	}
}

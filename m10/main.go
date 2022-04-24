package main

import (
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

/*
模块十作业（必交）
1、为 HTTPServer 添加 0-2 秒的随机延时；
2、为 HTTPServer 项目添加延时 Metric；
3、将 HTTPServer 部署至测试集群，并完成 Prometheus 配置；
4、从 Prometheus 界面中查询延时指标数据；
5、（可选）创建一个 Grafana Dashboard 展现延时分配情况。
提交地址： https://jinshuju.net/f/awEgbi
截止日期：2022 年 4 月 24 日 23:59
*/

/*
A:

1/2、见代码
	程序会自动请求自己的 hello 接口，以产生测试数据

3、deployment 部署
	deploy.yaml

4、查询 99 分位耗时
	histogram_quantile(0.99, sum(rate(httpserver_http_cost_bucket[5m])) by (le))

5、dashboard 配置，0.5、0.8、0.9、0.95 分位耗时分布
	dashboard/http-cost.json
*/

var Log *logrus.Logger

func init() {
	Log = logrus.New()
	Log.SetLevel(logrus.TraceLevel)
	Log.SetReportCaller(true)
	pathMap := lfshook.PathMap{
		logrus.InfoLevel:  "info.log",
		logrus.ErrorLevel: "error.log",
	}
	Log.Hooks.Add(lfshook.NewHook(
		pathMap,
		&logrus.JSONFormatter{},
	))
}

func main() {
	Log.Info("hello")
	preStop()

	initMetric()

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthz)
	mux.HandleFunc("/hello", hello)

	mux.Handle("/metrics", promhttp.Handler())

	// 自动请求自己，用于产生测试数据
	go func() {
		for {
			time.Sleep(time.Second)
			go func() {
				_, err := http.Get("http://localhost/hello")
				if err != nil {
					Log.Errorf("get hello err = %v", err)
				}
			}()
		}
	}()

	Log.Fatal(http.ListenAndServe(":80", mux))

	return
}

func healthz(w http.ResponseWriter, r *http.Request) {
	now := "[" + time.Now().String() + "] "
	Log.Info(now + "healthz")
	Log.Info("clientIP = ", getClientIP(r))

	Log.Info("http header")
	for k, vs := range r.Header {
		Log.Infof("%v: %v\n", k, vs)
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}

	w.Header().Set("VERSION", os.Getenv("VERSION"))

	_, _ = w.Write([]byte(now + "ok\n"))
}

func hello(w http.ResponseWriter, r *http.Request) {
	beg := time.Now()
	defer func() {
		setCost(beg)
		Log.Infof("hello, cost = %v", time.Now().Sub(beg).Milliseconds())
	}()

	delay := rand.Intn(2000)
	time.Sleep(time.Duration(delay) * time.Millisecond)

	_, err := io.WriteString(w, fmt.Sprintf("hello [%v]\n", delay))
	if err != nil {
		Log.Errorf("hello err = %v", err)
	}
}

func getClientIP(r *http.Request) string {
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	ip := strings.TrimSpace(strings.Split(xForwardedFor, ",")[0])
	if ip != "" {
		return ip
	}
	ip = strings.TrimSpace(r.Header.Get("X-Real-Ip"))
	if ip != "" {
		return ip
	}
	if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		return ip
	}
	return ""
}

func preStop() {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	go func() {
		s := <-signalChannel
		Log.Info("stop, signal: ", s)
		Log.Exit(0)
	}()
}

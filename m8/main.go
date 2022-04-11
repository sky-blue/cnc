package main

import (
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

//作业要求：编写 Kubernetes 部署脚本将 httpserver 部署到 Kubernetes 集群，以下是你可以思考的维度。
//
//优雅启动
//优雅终止
//资源需求和 QoS 保证
//探活
//日常运维需求，日志等级
//配置和代码分离
//[strong_begin] 提交地址： https://jinshuju.net/f/rJC4DG
//截止日期：2022 年 4 月 10 日 23:59

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

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
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
	})

	Log.Fatal(http.ListenAndServe(":80", mux))

	return
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

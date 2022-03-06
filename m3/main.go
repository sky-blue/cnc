package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

//本周作业
//编写一个 HTTP 服务器，大家视个人不同情况决定完成到哪个环节，但尽量把 1 都做完：
//
//接收客户端 request，并将 request 中带的 header 写入 response header
//读取当前系统的环境变量中的 VERSION 配置，并写入 response header
//Server 端记录访问日志包括客户端 IP，HTTP 返回码，输出到 server 端的标准输出
//当访问 localhost/healthz 时，应返回 200
//提交地址： https://jinshuju.net/f/eSZcZA
//截止日期：2022 年 2 月 27 日 23:59

func main() {
	fmt.Println("hello")

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		now := "[" + time.Now().String() + "] "
		fmt.Println(now + "healthz")
		fmt.Println("clentIP = ", getClientIP(r))

		fmt.Println("http header")
		for k, vs := range r.Header {
			fmt.Printf("%v: %v\n", k, vs)
			for _, v := range vs {
				w.Header().Add(k, v)
			}
		}

		w.Header().Set("VERSION", os.Getenv("VERSION"))

		w.Write([]byte(now + "ok\n"))
	})

	fmt.Println(http.ListenAndServe(":80", mux))

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

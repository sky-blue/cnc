FROM golang:1.17 AS build
WORKDIR /httpserver/
COPY . .

ENV CGO_ENABLED=0
ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn,direct

RUN GOOS=linux go build -installsuffix cgo -o httpserver


FROM busybox
COPY --from=build /httpserver/httpserver /httpserver/httpserver

ENV VERSION=1.2.3

WORKDIR /httpserver/
ENTRYPOINT ["./httpserver"]

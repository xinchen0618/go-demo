FROM golang:1.22.2-alpine3.19 AS build

ENV GOPROXY=https://goproxy.cn,direct
ENV CGO_ENABLED=0

WORKDIR /go-demo

# go.mod与其他文件分开处理是为了缓存, 并非每次更改代码都会导致依赖包变化
COPY go.mod go.sum ./

RUN go mod download

COPY . .

WORKDIR /go-demo/cmd/demo-api

RUN go build -ldflags="-s -w"


FROM alpine:3.19

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk add --update --no-cache tzdata
ENV TZ Asia/Shanghai

COPY --from=build /go-demo/cmd/demo-api/demo-api /

ENTRYPOINT ["/demo-api"]
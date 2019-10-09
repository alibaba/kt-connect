FROM registry.cn-hangzhou.aliyuncs.com/rdc-incubator/builder as build-deps
LABEL MAINTAINER yunlong <zhenmu.zyl@alibaba-inc.com>
ENV GO111MODULE on

WORKDIR /go/src/github.com/alibaba/kt-connect
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o output/kt-apiserver-linux-amd64 cmd/server/main.go 

FROM alpine
COPY --from=build-deps /go/src/github.com/alibaba/kt-connect/output/kt-apiserver-linux-amd64 /usr/local/bin/kt-apiserver
LABEL MAINTAINER yunlong <zhenmu.zyl@alibaba-inc.com>

CMD ["kt-apiserver"]
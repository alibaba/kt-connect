FROM registry.cn-hangzhou.aliyuncs.com/rdc-incubator/builder
ENV GO111MODULE on
WORKDIR /go/src/github.com/alibaba/kt-connect/
COPY . .
RUN GOARCH=amd64 GOOS=linux go build -o bin/kt-shadow cmd/proxy/main.go

FROM registry.cn-hangzhou.aliyuncs.com/rdc-incubator/shadow-base
COPY --from=0 /go/src/github.com/alibaba/kt-connect/bin/kt-shadow /usr/sbin/kt-shadow
CMD ["/usr/sbin/kt-shadow"]
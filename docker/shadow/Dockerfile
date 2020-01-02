FROM registry.cn-hangzhou.aliyuncs.com/rdc-incubator/builder
ENV GO111MODULE on
WORKDIR /go/src/github.com/alibaba/kt-connect/
COPY . .
RUN GOARCH=amd64 GOOS=linux go build -gcflags "all=-N -l" -o bin/kt-shadow cmd/shadow/main.go

FROM registry.cn-hangzhou.aliyuncs.com/rdc-incubator/shadow-base
COPY --from=0 /go/src/github.com/alibaba/kt-connect/bin/kt-shadow /usr/sbin/kt-shadow
RUN apt-get install -y net-tools
ADD docker/shadow/run.sh /run.sh
RUN chmod 755 /run.sh

CMD ["/run.sh"]
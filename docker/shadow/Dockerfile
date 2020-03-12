FROM golang:1.13
ENV GO111MODULE on
# Install go debugger
RUN go get -u github.com/go-delve/delve/cmd/dlv

FROM registry.cn-hangzhou.aliyuncs.com/rdc-incubator/shadow-base
COPY artifacts/shadow/shadow-linux-amd64 /usr/sbin/shadow-linux-amd64
COPY --from=0 /go/bin/dlv /usr/sbin/dlv
RUN apt-get install -y net-tools
ADD docker/shadow/run.sh /run.sh
RUN chmod 755 /run.sh

ENTRYPOINT ["/run.sh"]
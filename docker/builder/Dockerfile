FROM golang:1.13
LABEL MAINTAINER yunlong <zhenmu.zyl@alibaba-inc.com>
ENV GO111MODULE on

# Install go debugger
RUN go get -u github.com/go-delve/delve/cmd/dlv

# Install project dependences
WORKDIR /go/src/github.com/alibaba/kt-connect
COPY go.mod .
COPY go.sum .
RUN go mod download

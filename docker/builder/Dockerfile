FROM golang:1.11
LABEL MAINTAINER yunlong <zhenmu.zyl@alibaba-inc.com>
WORKDIR /go/src/github.com/alibaba/kt-connect

ENV GO111MODULE on

COPY go.mod .
COPY go.sum .

RUN go mod download
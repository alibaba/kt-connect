FROM golang:1.11
ENV GO111MODULE on
WORKDIR /go/src/github.com/alibaba/kt-connect/
COPY . .
RUN go mod download
RUN GOARCH=amd64 GOOS=linux go build -o "output/ktctl/ktctl-linux-amd64" ./cmd/ktctl
RUN GOARCH=386 GOOS=linux go build -o "output/ktctl/ktctl-linux-386" ./cmd/ktctl
RUN GOARCH=amd64 GOOS=darwin go build -o "output/ktctl/ktctl-darwin-amd64" ./cmd/ktctl
RUN GOARCH=386 GOOS=darwin go build -o "output/ktctl/ktctl-darwin-386" ./cmd/ktctl
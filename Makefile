
PREFIX			  ?= registry.cn-hangzhou.aliyuncs.com/rdc-incubator
TAG				  ?= $(shell date +%s)
SHADOW_IMAGE	  =  kt-connect-shadow
SHADOW_BASE_IMAGE =  shadow-base
BUILDER_IMAGE	  =  builder
DASHBOARD_IMAGE   =  kt-connect-dashboard
SERVER_IMAGE	  =  kt-connect-server

# generate mock
generate-mock:
	mockgen -source=pkg/kt/command/types.go -destination=pkg/mockd/mock/action_mock.go -package=mock
	mockgen -source=pkg/kt/cluster/types.go -destination=pkg/mockd/mock/kubernetes_mock.go -package=mock
	mockgen -source=pkg/kt/connect/types.go -destination=pkg/mockd/mock/connect_mock.go -package=mock

# run unit test
unit-test:
	mkdir -p artifacts/report/coverage
	go test -v -json -cover -coverprofile c.out.tmp ./...
	cat c.out.tmp | grep -v "_mock.go" > c.out

# build kt project
build: build-connect build-shadow build-server build-dashboard

# build ktctl
build-connect:
	bin/build-ktctl
	bin/archive

# build this image before shadow
build-shadow-base:
	docker build -t $(PREFIX)/$(SHADOW_BASE_IMAGE):$(TAG) -f docker/shadow/Dockerfile_base .

# build shadow
build-shadow:
	GOARCH=amd64 GOOS=linux go build -gcflags "all=-N -l" -o artifacts/shadow/shadow-linux-amd64 cmd/shadow/main.go
	docker build -t $(PREFIX)/$(SHADOW_IMAGE):$(TAG) -f docker/shadow/Dockerfile .
	docker push $(PREFIX)/$(SHADOW_IMAGE):$(TAG)

# dlv for debug
build-shadow-dlv:
	bin/build-shadow-dlv

build-dashboard:
	docker build -t $(PREFIX)/$(DASHBOARD_IMAGE):$(TAG) -f docker/dashboard/Dockerfile .
	docker push $(PREFIX)/$(DASHBOARD_IMAGE):$(TAG)

build-server:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o artifacts/apiserver/apiserver-linux-amd64 cmd/server/main.go
	docker build -t $(PREFIX)/$(SERVER_IMAGE):$(TAG) -f docker/apiserver/Dockerfile .
	docker push $(PREFIX)/$(SERVER_IMAGE):$(TAG)

git-release:
	bin/release
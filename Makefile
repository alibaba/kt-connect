
PREFIX			  ?= registry.cn-hangzhou.aliyuncs.com/rdc-incubator
TAG				  ?= $(shell date +%s)
SHADOW_IMAGE	  =  kt-connect-shadow
SHADOW_BASE_IMAGE =  shadow-base
BUILDER_IMAGE	  =  builder
DASHBOARD_IMAGE   =  kt-connect-dashboard
SERVER_IMAGE	  =  kt-connect-server

# generate mock
generate-mock:
	mkdir -p fake/kt/cluster
	mockgen -source=pkg/kt/cluster/types.go -destination=fake/kt/cluster/kubernetes_mock.go -package=cluster
	mkdir -p fake/kt/exec
	mockgen -source=pkg/kt/exec/kubectl/types.go -destination=fake/kt/exec/kubectl/kubectl_mock.go -package=kubectl
	mockgen -source=pkg/kt/exec/ssh/types.go -destination=fake/kt/exec/ssh/ssh_mock.go -package=ssh
	mockgen -source=pkg/kt/exec/sshuttle/types.go -destination=fake/kt/exec/sshuttle/sshuttle_mock.go -package=sshuttle
	mockgen -source=pkg/kt/exec/types.go -destination=fake/kt/exec/exec_mock.go -package=exec
	mockgen -source=pkg/kt/types.go -destination=fake/kt/kt_mock.go -package=kt

	mockgen -source=pkg/kt/channel/types.go -destination=pkg/kt/channel/mock.go -package=channel
    mockgen -source=pkg/kt/command/types.go -destination=pkg/kt/command/mock.go -package=command
    mockgen -source=pkg/kt/connect/types.go -destination=pkg/kt/connect/mock.go -package=connect

# run unit test
test:
	mkdir -p artifacts/report/coverage
	go test -v -cover -coverprofile c.out.tmp ./...
	cat c.out.tmp | grep -v "_mock.go" > c.out
	go tool cover -html=c.out -o artifacts/report/coverage/index.html

# build kt project
build: build-connect build-shadow build-server build-dashboard

# check the style
check:
	golint ./pkg/... ./cmd/...
# 	golangci-lint run ./pkg/...

# build ktctl
build-connect:
	scripts/build-ktctl
	scripts/archive

# build connect plugin
build-connect-plugin:
	scripts/build-kubectl-plugin-connect
	scripts/archive-plugins

# build this image before shadow
build-shadow-base:
	docker build -t $(PREFIX)/$(SHADOW_BASE_IMAGE):$(TAG) -f build/docker/shadow/Dockerfile_base .

# build shadow
build-shadow:
	GOARCH=amd64 GOOS=linux go build -gcflags "all=-N -l" -o artifacts/shadow/shadow-linux-amd64 cmd/shadow/main.go
	docker build -t $(PREFIX)/$(SHADOW_IMAGE):$(TAG) -f build/docker/shadow/Dockerfile .

# release shadow
release-shadow:
	docker push $(PREFIX)/$(SHADOW_IMAGE):$(TAG)

# dlv for debug
build-shadow-dlv:
	scripts/build-shadow-dlv

build-dashboard: build-frontend build-server

build-frontend:
	docker build -t $(PREFIX)/$(DASHBOARD_IMAGE):$(TAG) -f build/docker/dashboard/Dockerfile .

build-server:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o artifacts/apiserver/apiserver-linux-amd64 cmd/server/main.go
	docker build -t $(PREFIX)/$(SERVER_IMAGE):$(TAG) -f build/docker/apiserver/Dockerfile .

release-dashboard:
	docker push $(PREFIX)/$(DASHBOARD_IMAGE):$(TAG)
	docker push $(PREFIX)/$(SERVER_IMAGE):$(TAG)

git-release:
	scripts/release

dryrun-release:
	goreleaser --snapshot --skip-publish --rm-dist

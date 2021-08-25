
PREFIX			  ?= registry.cn-hangzhou.aliyuncs.com/rdc-incubator
TAG				  ?= $(shell date +%s)
SHADOW_IMAGE	  =  kt-connect-shadow
SHADOW_BASE_IMAGE =  shadow-base
BUILDER_IMAGE	  =  builder
DASHBOARD_IMAGE   =  kt-connect-dashboard
SERVER_IMAGE	  =  kt-connect-server

# generate mock
generate-mock:
	echo "generate mocks"
	mockgen -source=pkg/kt/cluster/types.go -destination=pkg/kt/cluster/mock.go -package=cluster
	mockgen -source=pkg/kt/command/types.go -destination=pkg/kt/command/mock.go -package=command
	mockgen -source=pkg/kt/connect/types.go -destination=pkg/kt/connect/mock.go -package=connect
	mockgen -source=pkg/kt/exec/kubectl/types.go -destination=pkg/kt/exec/kubectl/mock.go -package=kubectl
	mockgen -source=pkg/kt/exec/portforward/types.go -destination=pkg/kt/exec/portforward/mock.go -package=portforward
	mockgen -source=pkg/kt/exec/ssh/types.go -destination=pkg/kt/exec/ssh/mock.go -package=ssh
	mockgen -source=pkg/kt/exec/sshchannel/types.go -destination=pkg/kt/exec/sshchannel/mock.go -package=sshchannel
	mockgen -source=pkg/kt/exec/sshuttle/types.go -destination=pkg/kt/exec/sshuttle/mock.go -package=sshuttle
	mockgen -source=pkg/kt/exec/types.go -destination=pkg/kt/exec/mock.go -package=exec
	mockgen -source=pkg/kt/exec/tunnel/types.go -destination=pkg/kt/exec/tunnel/mock.go -package=tunnel
	mockgen -source=pkg/kt/types.go -destination=pkg/kt/mock.go -package=kt

# run unit test
test:
	mkdir -p artifacts/report/coverage
	go test -v -cover -coverprofile c.out.tmp ./...
	cat c.out.tmp | grep -v "_mock.go" > c.out
	go tool cover -html=c.out -o artifacts/report/coverage/index.html

# build kt project
compile:
	goreleaser --snapshot --skip-publish --rm-dist

# check the style
check:
	golint ./pkg/... ./cmd/...
# 	golangci-lint run ./pkg/...

# build this image before shadow
build-shadow-base:
	docker build -t $(PREFIX)/$(SHADOW_BASE_IMAGE):$(TAG) -f build/docker/shadow/Dockerfile_base .

# build shadow
build-shadow:
	GOARCH=amd64 GOOS=linux go build -gcflags "all=-N -l" -o artifacts/shadow/shadow-linux-amd64 cmd/shadow/main.go
	docker build -t $(PREFIX)/$(SHADOW_IMAGE):$(TAG) -f build/docker/shadow/Dockerfile .

# build router
build-router:
    GOARCH=amd64 GOOS=linux go build -gcflags "all=-N -l" -o artifacts/router/router-linux-amd64 cmd/router/main.go

# dlv for debug
build-shadow-dlv:
	make build-shadow TAG=latest
	scripts/build-shadow-dlv

build-dashboard: build-frontend

build-frontend:
	docker build -t $(PREFIX)/$(DASHBOARD_IMAGE):$(TAG) -f build/docker/dashboard/Dockerfile .

release-dashboard:
	docker push $(PREFIX)/$(DASHBOARD_IMAGE):$(TAG)
	docker push $(PREFIX)/$(SERVER_IMAGE):$(TAG)

clean:
	rm -fr artifacts

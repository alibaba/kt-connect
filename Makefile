PREFIX			  ?= registry.cn-hangzhou.aliyuncs.com/rdc-incubator
TAG				  ?= $(shell date +%s)
SHADOW_IMAGE	  =  kt-connect-shadow
SHADOW_BASE_IMAGE =  shadow-base
ROUTER_IMAGE	  =  kt-connect-router

# generate mock
generate-mock:
	echo "generate mocks"
	mockgen -source=pkg/kt/cluster/types.go -destination=pkg/kt/cluster/mock.go -package=cluster
	mockgen -source=pkg/kt/command/types.go -destination=pkg/kt/command/mock.go -package=command
	mockgen -source=pkg/kt/exec/sshchannel/types.go -destination=pkg/kt/exec/sshchannel/mock.go -package=sshchannel
	mockgen -source=pkg/kt/exec/sshuttle/types.go -destination=pkg/kt/exec/sshuttle/mock.go -package=sshuttle
	mockgen -source=pkg/kt/exec/types.go -destination=pkg/kt/exec/mock.go -package=exec
	mockgen -source=pkg/kt/exec/tun/types.go -destination=pkg/kt/exec/tun/mock.go -package=tun
	mockgen -source=pkg/kt/types.go -destination=pkg/kt/mock.go -package=kt

# run mod tidy
mod:
	go mod tidy -compat=1.17

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
	go vet ./pkg/... ./cmd/...

# build ktctl
build-ktctl:
	GOARCH=amd64 GOOS=linux go build -o "artifacts/ktctl/ktctl-linux" ./cmd/ktctl
	GOARCH=amd64 GOOS=darwin go build -o "artifacts/ktctl/ktctl-darwin" ./cmd/ktctl
	GOARCH=amd64 GOOS=windows go build -o "artifacts/ktctl/ktctl-windows" ./cmd/ktctl

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
	docker build -t $(PREFIX)/$(ROUTER_IMAGE):$(TAG) -f build/docker/router/Dockerfile .

# dlv for debug
build-shadow-dlv:
	make build-shadow TAG=latest
	scripts/build-shadow-dlv

clean:
	rm -fr artifacts

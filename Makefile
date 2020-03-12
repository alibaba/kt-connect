
PREFIX			  ?= registry.cn-hangzhou.aliyuncs.com/rdc-incubator
TAG				  ?= latest
SHADOW_IMAGE	  =  kt-connect-shadow
SHADOW_BASE_IMAGE =  shadow-base
BUILDER_IMAGE	  =  builder
DASHBOARD_IMAGE   =  kt-connect-dashboard
SERVER_IMAGE	  =  kt-connect-server

# run unit test
unit-test:
	mkdir -p artifacts/report/coverage
	go test -v -json -cover -coverprofile artifacts/report/coverage/cover.out ./...
	go tool cover -html=artifacts/report/coverage/cover.out -o artifacts/report/coverage/index.html

# build kt project
build: build-connect build-shadow

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

# build this first,it's the base image 
build-builder:
	docker build -t $(PREFIX)/$(BUILDER_IMAGE):$(TAG) -f docker/builder/Dockerfile .

# dlv for debug
build-shadow-dlv:
	bin/build-shadow-dlv

build-dashboard:
	docker build -t $(PREFIX)/$(DASHBOARD_IMAGE):$(TAG) -f docker/dashboard/Dockerfile . && \

build-server:
	docker build -t $(PREFIX)/$(SERVER_IMAGE):$(TAG) -f docker/apiserver/Dockerfile .

release-docker: build-builder build-shadow-base build-shadow build-connect build-dashboard build-server
	docker push $(PREFIX)/$(SHADOW_IMAGE)	
	# todo: push as you want 

git-release:
	bin/release
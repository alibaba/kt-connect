
PREFIX			  ?= registry.cn-hangzhou.aliyuncs.com/rdc-incubator
TAG				  ?= latest
SHADOW_IMAGE	  =  kt-connect-shadow
SHADOW_BASE_IMAGE =  shadow-base
CONNECT_IMAGE	  =  kt-connect
BUILDER_IMAGE	  =  builder
DASHBOARD_IMAGE   =  kt-connect-dashboard
SERVER_IMAGE	  =  kt-connect-server

# run unit test
test:
	go test ./...

# build this first,it's the base image 
build-builder:
	docker build -t $(PREFIX)/$(BUILDER_IMAGE):$(TAG) -f docker/builder/Dockerfile .

# build this image before shadow
build-shadow-base:
	docker build -t $(PREFIX)/$(SHADOW_BASE_IMAGE):$(TAG) -f docker/shadow/Dockerfile_base .

build-shadow:
	docker build -t $(PREFIX)/$(SHADOW_IMAGE):$(TAG) -f docker/shadow/Dockerfile .

# dlv for debug
build-shadow-dlv:
	bin/build-shadow-dlv

build-connect:
	bin/build-ktctl
	bin/archive

build-dashboard:
	docker build -t $(PREFIX)/$(DASHBOARD_IMAGE):$(TAG) -f docker/dashboard/Dockerfile . && \

build-server:
	docker build -t $(PREFIX)/$(SERVER_IMAGE):$(TAG) -f docker/apiserver/Dockerfile .

release-docker: build-builder build-shadow-base build-shadow build-connect build-dashboard build-server
	docker push $(PREFIX)/$(SHADOW_IMAGE)	
	# todo: push as you want 

git-release:
	bin/release
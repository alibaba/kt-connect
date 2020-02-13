KT Connect
===========

[![Build Status](https://travis-ci.org/alibaba/kt-connect.svg?branch=master)](https://travis-ci.org/alibaba/kt-connect) ![License](https://img.shields.io/github/license/alibaba/kt-connect.svg) 

![Logo](./docs/media/logo.png)

Management and Integration with your Kubernetes dev environment more efficient.

![Arch](./docs/media/arch.png)

## Features

* Connect: Direct access remote Kubernetes cluster: KT Connect use `sshuttle` as the vpn tool to access remote Kubernetes cluster network.
* Exchange: Developer can exchange the workload to redirect the request to local app.
* Mesh: You can create a mesh version in local host and redirect to your local
* Dashboard: A Dashboard view you can know how the environment is use.

## QuickStart

You can download and install the ktctl from [Downloads And Install](https://rdc-incubator.github.io/kt-docs/#/downloads)

### Deploy a service in Kubernetes

```
$ kubectl run tomcat --image=tomcat:9 --expose --port=8080
kubectl run --generator=deployment/apps.v1 is DEPRECATED and will be removed in a future version. Use kubectl run --generator=run-pod/v1 or kubectl create instead.
service/tomcat created
deployment.apps/tomcat created

# Deployment info
$ kubectl get deployments -o wide --selector run=tomcat
NAME     DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE   CONTAINERS   IMAGES     SELECTOR
tomcat   1         1         1            1           12m   tomcat       tomcat:9   run=tomcat

# Pods info
$ kubectl get pods -o wide --selector run=tomcat
NAME                     READY   STATUS        RESTARTS   AGE   IP             NODE                                NOMINATED NODE
tomcat-cc7648444-r9tw4   1/1     Running       0          2m    172.16.0.147   cn-beijing.i-2ze11lz4lijf1pmecnwp   <none>

# Service info
$ kubectl get svc tomcat
NAME     TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)    AGE
tomcat   ClusterIP   172.19.143.139   <none>        8080/TCP   4m
```

### Connect: Access Kubernetes resource from local

```
$ sudo ktctl connect
2019/06/19 11:11:07 Deploying proxy deployment kt-connect-daemon in namespace default
2019/06/19 11:11:07 Pod status is Pending
2019/06/19 11:11:09 Pod status is Running
2019/06/19 11:11:09 Success deploy proxy deployment kt-connect-daemon in namespace default
2019/06/19 11:11:18 KT proxy start successful
```

Access PodIP:

```
curl http://172.16.0.147:8080 
```

Access ClusterIP:

```
$ curl http://172.19.143.139:8080
```

Access Server internal DNS address

```
$ curl http://tomcat:8080
```

### Exchange: Access local from cluster

Create Tomcat 8 in local and expose 8080 port

```
docker run -itd -p 8080:8080 tomcat:8
```

```
$ sudo ktctl exchange tomcat --expose 8080
2019/06/19 11:19:10  * tomcat (0 replicas)
2019/06/19 11:19:10 Scale deployment tomcat to zero
2019/06/19 11:19:10 Deploying proxy deployment tomcat-kt-oxpjf in namespace default
2019/06/19 11:19:10 Pod status is Pending
2019/06/19 11:19:12 Pod status is Running
2019/06/19 11:19:12 Success deploy proxy deployment tomcat-kt-oxpjf in namespace default
SSH Remote port-forward for POD starting
2019/06/19 11:19:14 ssh remote port-forward start at pid: 3567
```

```
tips:
if your remote port diffrent local port, you can user ${remotePort}:{localPort} set expose parameter
```

Access local tomcat by internal service DNS address:

> Note: if `kubectl connect` not running, you can only access from cluster

```
$ curl http://tomcat:8080 | grep '<h1>'
<h1>Apache Tomcat/8.5.37</h1> #
```

### Mesh: Build large test environemnt with ServiceMesh

> You can know more from [Mesh Best Practices](https://rdc-incubator.github.io/kt-docs/#/guide/mesh)

The most different from mesh and exchange is exchange will scale the origin workload replicas to zero. And messh will keep it and create a pod instance with random version, after this user can modifi the Istio route rule let the specific request redirect to local, and the environment is working as normal:

```
$ sudo ktctl mesh tomcat --expose 8080
2019/06/19 22:10:23 'KT Connect' not runing, you can only access local app from cluster
2019/06/19 22:10:24 Deploying proxy deployment tomcat-kt-ybocr in namespace default
2019/06/19 22:10:24 Pod status is Pending
2019/06/19 22:10:26 Pod status is Pending
2019/06/19 22:10:28 Pod status is Running
2019/06/19 22:10:28 Success deploy proxy deployment tomcat-kt-ybocr in namespace default
2019/06/19 22:10:28 -----------------------------------------------------------
2019/06/19 22:10:28 |    Mesh Version 'ybocr' You can update Istio rule       |
2019/06/19 22:10:28 -----------------------------------------------------------
2019/06/19 22:10:30 exchange port forward to local start at pid: 53173
SSH Remote port-forward POD 172.16.0.217 22 to 127.0.0.1:2217 starting
2019/06/19 22:10:30 ssh remote port-forward exited
2019/06/19 22:10:32 ssh remote port-forward start at pid: 53174
```

### Dashboard

Dashboard can help your know how your dev environemnt is used. 

![Dashboard](./docs/media/dashboard-demo.png)

You can install KT Connect Dashboard As Follow [Install Dashboard](https://rdc-incubator.github.io/kt-docs/#/guide/dashboard)

## How To Build

### use go build

```
go build -o "output/ktctl/ktctl" ./cmd/ktctl
```

### use gox

Install Gox

```
go get github.com/mitchellh/gox
```

```
gox -output="output/ktctl/ktctl_{{.OS}}_{{.Arch}}" ${GOPATH}/src/github.com/alibaba/kt-connect/cmd/ktctl
```

## How To Build Shadow

```
GOARCH=amd64 GOOS=linux go build -o bin/proxy cmd/shadow/main.go
```

## Ask For Help

Please contact us with DingTalk:

![image](./docs/media/dingtalk-group.png)

## Release Note

### 0.0.9

> Release At 2020-01-16

* Support Service Name as dns address
* Make sure shadow is clean up after command exit

### 0.0.8

> Release At 2019-12-13

* Add windows native support
* Add idea support

### 0.0.7

> Release At 2019-12-05

* Add oidc plugin to support TKE
* Add socks5 method to support WSL
* Fixed issue when node.Spec.PodCIRD dynamic cal CIRD

### 0.0.6

> Release At 2019-10-01

* Fixed clusterIP cidr missing.
* Rename docker images address.

### 0.0.5

> Release At 2019-10-09

* Add dashboard and api server source code

### 0.0.4

> Release At 2019-06-26

* Support KT Connect Dashboard

### 0.0.3

> Release At 2019-06-19

* Add `mesh` command to support istio network rule

### 0.0.2

> Release At 2019-06-19

* Fixed issue if istio inject is enable in namespace, and the request can't redirect to local
* Support exchange run standalone.

### 0.0.1 

> Release At 2019-06-18

* split command to `connect` and `exchange`.
* support mutil exchange.
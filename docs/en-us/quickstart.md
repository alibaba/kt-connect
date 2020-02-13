# Quick Start Guide

In this chapter, we will deployment a sample app(tomcat:9) in Kubernetes cluster. With Kt to access the app from user labtop or exchange the request to user labtop.

## Create a Demo APP in Cluster

``` shell
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

## Connect: Access APP From Local

KT Connect support `VPN mode` and `Socks5 mode` to access cluster from local:

<!-- tabs:start -->

#### ** VPN Mode **

> VPN Mode is base on sshuttle, only support Mac and Linux User

Connect to kubernetes cluster, KT will deployment a proxy pod in cluster：

```shell
$ sudo ktctl connect --method=vpn
2019/06/19 11:11:07 Deploying proxy deployment kt-connect-daemon in namespace default
2019/06/19 11:11:07 Pod status is Pending
2019/06/19 11:11:09 Pod status is Running
2019/06/19 11:11:09 Success deploy proxy deployment kt-connect-daemon in namespace default
2019/06/19 11:11:18 KT proxy start successful
```

```
$ curl http://172.16.0.147:8080 #Access PodIP from local
$ curl http://172.19.143.139:8080 #Access ClusterIP
$ curl http://tomcat:8080 #Access Server internal DNS address    
```

#### ** SOCKS5 Mode **

> Socks5 Mode is base SSH can support windows/linux/mac

use `--method=socks` tell kt connect use socks5 mode, and `--dump2hosts` option will sync all service in namespace to hosts file.


```
$ sudo ktctl connect --method=socks5 --dump2hosts
6:37PM INF Connect Start At 74032
6:37PM INF Dump hosts successful.
6:37PM INF Client address 30.5.124.242
6:37PM INF Deploying shadow deployment kt-connect-daemon-mkevz in namespace default
6:37PM INF Shadow is ready.
6:37PM INF Success deploy proxy deployment kt-connect-daemon-mkevz in namespace default

Forwarding from 127.0.0.1:2222 -> 22
Forwarding from [::1]:2222 -> 22
6:38PM INF ==============================================================
6:38PM INF Start SOCKS5 Proxy: export http_proxy=socks5://127.0.0.1:2223
6:38PM INF ==============================================================
Handling connection for 2222
Warning: Permanently added '[127.0.0.1]:2222' (ECDSA) to the list of known hosts.
6:38PM INF KT proxy start successful
```

Set HTTP_PROXY Environment in shell and then access cluster resource from local：

```
$ export http_proxy=socks5://127.0.0.1:2223
$ curl http://172.16.0.147:8080 #本地直接访问PodIP
$ curl http://172.19.143.139:8080 # 本地直接访问ClusterIP
$ curl http://tomcat:8080 #使用Service的域名访问
```

> when ktctl hosts file will auto clean

<!-- tabs:end -->


## Exchange: Access local from cluster

Create Tomcat 8 in local and expose 8080 port

```
$ docker run -itd -p 8080:8080 tomcat:8
```

```
$ ktctl exchange tomcat --expose 8080
2019/06/19 11:19:10  * tomcat (0 replicas)
2019/06/19 11:19:10 Scale deployment tomcat to zero
2019/06/19 11:19:10 Deploying proxy deployment tomcat-kt-oxpjf in namespace default
2019/06/19 11:19:10 Pod status is Pending
2019/06/19 11:19:12 Pod status is Running
2019/06/19 11:19:12 Success deploy proxy deployment tomcat-kt-oxpjf in namespace default
SSH Remote port-forward for POD starting
2019/06/19 11:19:14 ssh remote port-forward start at pid: 3567
```

Access local tomcat by internal service DNS address:

> Note: if `kubectl connect` not running, you can only access from cluster

```
$ curl http://tomcat:8080 | grep '<h1>'
<h1>Apache Tomcat/8.5.37</h1> #
```

## Mesh: Build large test environemnt with ServiceMesh

The most different from `mesh` and `exchange` is exchange will scale the origin workload replicas to zero. And messh will keep it and create a pod instance with random `version`, after this user can modifi the Istio route rule let the specific request redirect to local, and the environment is working as normal:

```
$ ktctl mesh tomcat --expose 8080
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
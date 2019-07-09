# 快速开始

这里，我们通过一个简单的示例，来体验KT的主要能力。 这里我们会在集群中部署一个Tomcat:7的镜像，并且通过KT直接在本地访问该服务，同时也可以将集群中所有原本对该应用的请求转发到本地。

## 部署实例应用

``` shell
$ kubectl run tomcat --image=tomcat:7 --expose --port=8080
kubectl run --generator=deployment/apps.v1 is DEPRECATED and will be removed in a future version. Use kubectl run --generator=run-pod/v1 or kubectl create instead.
service/tomcat created
deployment.apps/tomcat created

# Deployment info
$ kubectl get deployments -o wide --selector run=tomcat
NAME     DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE   CONTAINERS   IMAGES     SELECTOR
tomcat   1         1         1            1           12m   tomcat       tomcat:7   run=tomcat

# Pods info
$ kubectl get pods -o wide --selector run=tomcat
NAME                     READY   STATUS        RESTARTS   AGE   IP             NODE                                NOMINATED NODE
tomcat-cc7648444-r9tw4   1/1     Running       0          2m    172.16.0.147   cn-beijing.i-2ze11lz4lijf1pmecnwp   <none>

# Service info
$ kubectl get svc tomcat
NAME     TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)    AGE
tomcat   ClusterIP   172.19.143.139   <none>        8080/TCP   4m
```

## Connect：连接集群网络

使用connect命令建立本地到集群的VPN网络：

![](../_media/demo-1.gif)

```shell
$ sudo ktctl connect
2019/06/19 11:11:07 Deploying proxy deployment kt-connect-daemon in namespace default
2019/06/19 11:11:07 Pod status is Pending
2019/06/19 11:11:09 Pod status is Running
2019/06/19 11:11:09 Success deploy proxy deployment kt-connect-daemon in namespace default
2019/06/19 11:11:18 KT proxy start successful
```

在本地直接访问PodIP:

```
$ curl http://172.16.0.147:8080             
```

在本地访问ClusteriIP:

```
$ curl http://172.19.143.139:8080             
```

使用Service的域名访问：

```
$ curl http://tomcat.default.svc.cluster.local:8080
```

## Exchange: 将集群流量转发到本地

![](../_media/demo-2.gif)

为了模拟集群联调本地的情况，我们首先在本地运行一个Tomcat:8的容器

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

在本地或者集群中访问原本指向Tomcat:7的应用，查看输出结果：

> 注意如果未运行`ktctl connect`,只能从集群内访问

```
$ curl http://tomcat.default.svc.cluster.local:8080 | grep '<h1>'
<h1>Apache Tomcat/8.5.37</h1> #
```

## Mesh: 基于Service Mesh按规则转发流量到本地

> 查看更多：[Mesh最佳实践](/zh-cn/guide/mesh)

`mesh`与`exchange`的最大区别在于，exchange会完全替换原有的应用实例。mesh命令创建代理容器，但是会保留原应用容器，代理容器会动态生成version标签，以便用于可以通过Istio流量规则将特定的流量转发到本地，同时保证环境正常链路始终可用：

![](../_media/demo-3.gif)

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
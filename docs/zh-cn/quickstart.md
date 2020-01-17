# 快速开始

这里，我们通过一个简单的示例，来体验KT的主要能力。 这里我们会在集群中部署一个Tomcat:7的镜像，并且通过KT直接在本地访问该服务，同时也可以将集群中所有原本对该应用的请求转发到本地。

## 部署实例应用

``` shell
$ kubectl run tomcat --image=tomcat:7 --expose --port=8080
kubectl run --generator=deployment/apps.v1 is DEPRECATED and will be removed in a future version. Use kubectl run --generator=run-pod/v1 or kubectl create instead.
service/tomcat created
deployment.apps/tomcat created

$ kubectl get deployments -o wide --selector run=tomcat
NAME     DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE   CONTAINERS   IMAGES     SELECTOR
tomcat   1         1         1            1           12m   tomcat       tomcat:7   run=tomcat

$ kubectl get pods -o wide --selector run=tomcat
NAME                     READY   STATUS        RESTARTS   AGE   IP             NODE                                NOMINATED NODE
tomcat-cc7648444-r9tw4   1/1     Running       0          2m    172.16.0.147   cn-beijing.i-2ze11lz4lijf1pmecnwp   <none>

$ kubectl get svc tomcat
NAME     TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)    AGE
tomcat   ClusterIP   172.19.143.139   <none>        8080/TCP   4m
```

## Connect：连接集群网络

使用connect命令建立从本地到集群的网络通道，目前KT Connect支持VPN和Socks5代理两种模式, 不指定`--method`参数时，默认使用VPN模式：

<!-- tabs:start -->

#### ** VPN 模式 **

> VPN模式基于sshuttle目前只支持Mac/Linux环境

使用connect命令建立本地到集群的VPN网络：

```shell
$ sudo ktctl connect --method=vpn
2019/06/19 11:11:07 Deploying proxy deployment kt-connect-daemon in namespace default
2019/06/19 11:11:07 Pod status is Pending
2019/06/19 11:11:09 Pod status is Running
2019/06/19 11:11:09 Success deploy proxy deployment kt-connect-daemon in namespace default
2019/06/19 11:11:18 KT proxy start successful
```

启用VPN后直接访问集群资源：

```
$ curl http://172.16.0.147:8080      #在本地直接访问PodIP
$ curl http://172.19.143.139:8080   #在本地访问ClusteriIP
$ curl http://tomcat:8080     #使用Service的域名访问
```

#### ** Socks5代理模式 **

> Socks5模式基于ssh目前只支持Windows/Mac/Linux环境

使用`--method=socks5`指定使用socks5代理模式，为了能够在本地直接访问service的DNS域名`--dump2hosts`可以自动同步指定命名空间下的所有Service到本地的hosts文件：

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

在Shell中按照日志输出提示，设置http_proxy参数：

```
$ export http_proxy=socks5://127.0.0.1:2223
$ curl http://172.16.0.147:8080 #本地直接访问PodIP
$ curl http://172.19.143.139:8080 # 本地直接访问ClusterIP
$ curl http://tomcat:8080 #使用Service的域名访问
```

> 当命令退出后会自动清理本地的hosts文件

<!-- tabs:end -->

## Exchange: 将集群流量转发到本地

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
$ curl http://tomcat:8080 | grep '<h1>'
<h1>Apache Tomcat/8.5.37</h1> #
```

## Mesh: 基于Service Mesh按规则转发流量到本地

> 查看更多：[Mesh最佳实践](/zh-cn/guide/mesh)

`mesh`与`exchange`的最大区别在于，exchange会完全替换原有的应用实例。mesh命令创建代理容器，但是会保留原应用容器，代理容器会动态生成version标签，以便用于可以通过Istio流量规则将特定的流量转发到本地，同时保证环境正常链路始终可用：

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
快速开始
---

KtConnect提供了本地和测试环境集群的双向互联能力。在这篇文档里，我们将使用一个简单的示例，来快速演示通过KtConnect完成本地直接访问集群中的服务、以及将集群中指定服务的请求转发到本地的过程。

## 部署实例应用

为了便于展示结果，首先在集群中部署一个Tomcat服务并创建一个默认首页：

```bash
$ kubectl create deployment tomcat --image=tomcat:9 --port=8080
deployment.apps/tomcat created

$ kubectl expose deployment tomcat --port=8080 --target-port=8080
service/tomcat exposed

$ kubectl exec deployment/tomcat -c tomcat -- /bin/bash -c 'mkdir webapps/ROOT; echo "kt-connect demo v1" > webapps/ROOT/index.html'
```

查询Pod和服务的IP地址：

```bash
$ kubectl get pod -o wide --selector app=tomcat
NAME     READY   STATUS    RESTARTS   AGE   IP            ...
tomcat   1/1     Running   0          34s   10.51.0.162   ...

$ kubectl get svc tomcat
NAME     TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)    AGE
tomcat   ClusterIP   172.16.255.111   <none>        8080/TCP   34s
```

可知Tomcat实例的Pod IP为`10.51.0.162`，服务的Cluster IP为`172.16.255.111`，记下待用。

## 连接集群网络

使用`ktctl connect`命令建立从本地到集群的网络通道，注意该命令需要管理员权限。

<!-- tabs:start -->

#### ** MacOS/Linux **

在Mac/Linux下可通过`sudo`执行：

```bash
$ sudo ktctl connect
00:00AM INF KtConnect start at <PID>
... ...
00:00AM INF ---------------------------------------------------------------
00:00AM INF  All looks good, now you can access to resources in the kubernetes cluster
00:00AM INF ---------------------------------------------------------------
```

#### ** Windows **

在Windows下可以在CMD或PowerShell图标上右键，点击"以管理员身份运行"，然后在打开的窗口里执行：

```bash
> ktctl connect
00:00AM INF KtConnect start at <PID>
... ...
00:00AM INF ---------------------------------------------------------------
00:00AM INF  All looks good, now you can access to resources in the kubernetes cluster
00:00AM INF ---------------------------------------------------------------
```

<!-- tabs:end -->

现在本地已经能够直接访问集群资源了，可通过浏览器或`curl`命令来验证：

> 注意：在Windows PowerShell中，`curl`是一个内置命令，输出格式与下述示例有所不同，可用`curl.exe`替代命令中的`curl`

```bash
$ curl http://10.51.0.162:8080    # 在本地直接访问PodIP
kt-connect demo v1

$ curl http://172.21.6.39:8080    # 在本地访问ClusterIP
kt-connect demo v1

$ curl http://tomcat:8080         # 使用<service>作为域名访问服务
kt-connect demo v1

$ curl http://tomcat.default:8080     # 使用<servicename>.<namespace>域名访问服务
kt-connect demo v1

$ curl http://tomcat.default.svc.cluster.local:8080    # 使用集群内完整域名访问服务
kt-connect demo v1
```

## 将集群流量转发到本地

为了验证集群访问本地服务的场景，我们在本地也启动一个Tomcat的容器，并为其创建一个内容不同的首页。

```bash
$ docker run -d --name tomcat -p 8080:8080 tomcat:9
$ docker exec tomcat /bin/bash -c 'mkdir webapps/ROOT; echo "kt-connect local v2" > webapps/ROOT/index.html'
```

KtConnect提供了两种能够让集群流量重定向到本地服务的命令，在使用场景上稍有不同。

- Exchange：将集群指定服务的所有流量转向本地
- Mesh：将集群指定服务的部分流量（按Header或Label规则）转向本地

<!-- tabs:start -->

#### ** Exchange命令 **

将集群里访问指定服务的所有请求拦截并转发到本地的指定端口。通常用于调试在测试环境里，调试位于业务调用链中间环节的特定服务。

```text
┌──────────┐     ┌─ ── ── ──     ┌──────────┐
│ ServiceA ├─┬─►x│ ServiceB │ ┌─►│ ServiceC │
└──────────┘ │    ── ── ── ─┘ │  └──────────┘
         exchange             │
             │   ┌──────────┐ │
             └──►│ ServiceB'├─┘ （本地服务实例）
                 └──────────┘
```

使用`ktctl exchange`命令将先前部署到集群中的`tomcat`服务流量全部转到本地`8080`端口：

```bash
$ ktctl exchange tomcat --expose 8080
00:00AM INF KtConnect start at <PID>
... ...
---------------------------------------------------------------
 Now all request to service 'tomcat' will be redirected to local
---------------------------------------------------------------
```

在本地或者集群中访问示例开始时部署到集群的`tomcat`服务，查看输出结果：

> 注意如果未运行`ktctl connect`，只能从集群内访问

```bash
$ curl http://tomcat:8080
kt-connect local v2
```

可以看到，访问集群里`tomcat`服务的请求被路由到了本地的Tomcat实例，现在就可以直接在本地调试这个服务了。

## ** Mesh命令 **

将集群里访问指定服务的部分请求拦截并转发到本地的指定端口。通常用于团队协作时，需要定向调试调用链中间位置的服务，又不希望影响其他开发者正常使用测试环境的场景。

```text
┌──────────┐     ┌──────────┐    ┌──────────┐
│ ServiceA ├─┬──►│ ServiceB │─┬─►│ ServiceC │
└──────────┘ │   └──────────┘ │  └──────────┘
            mesh              │
             │   ┌──────────┐ │
             └──►│ ServiceB'├─┘ （本地服务实例）
                 └──────────┘
```

Mesh命令有两种运行模式，默认的`auto`模式**不需要**额外的服务网格组件，能够直接实现HTTP请求的自动按需路由。

为了便于验证结果，先重置一下集群里Tomcat服务的首页内容。然后通过`ktctl mesh`命令创建代理Pod：

```bash
$ kubectl exec deployment/tomcat -c tomcat -- /bin/bash -c 'mkdir webapps/ROOT; echo "kt-connect demo v1" > webapps/ROOT/index.html'

$ ktctl mesh tomcat --expose 8080
00:00AM INF KtConnect start at <PID>
... ...
--------------------------------------------------------------
 Now you can access your service by header 'VERSION: feo3x'
--------------------------------------------------------------
```

在命令日志的末尾，输出了一个特定的Header值。此时，直接访问集群里的`tomcat`服务，流量将正常进入集群的服务实例：

```bash
$ curl http://tomcat:8080
kt-connect demo v1
```

若请求包含Mesh命令输出的Header，则流量将自动被本地的服务实例接收。

```bash
$ curl -H 'VERSION: feo3x' http://tomcat:8080
kt-connect local v2
```

在实际使用时，可结合[ModHeader插件](https://github.com/bewisse/modheader)，使得只有开发者从自己浏览器发出的请求会访问其本地的服务进程。

除此以外，还有一种可灵活配置路由规则的`manual`模式，该模式下KtConnect不会自动创建路由，在Mesh命令运行后，访问指定服务的流量将随机访问集群服务和本地实例。您可以自行使用任何服务网格组件（譬如Istio）创建基于`version`标签的路由规则，将特定流量转发到本地。详见[Manual Mesh](../reference/manual_mesh.md)文档。

`ktctl exchange`与`ktctl mesh`命令的最大区别在于，前者会将原应用实例流量全部替换为由本地服务接收，而后者仅将包含指定Header的流量导流到本地，同时保证测试环境正常链路始终可用。

<!-- tabs:end -->

## 将本地服务提供给其他开发者

除了已经部署到集群的服务，在开发过程中，也可以利用KtConnect将本地服务快速"放"到集群，变成一个临时的服务，供其他开发者或集群中的其他服务使用。

- Preview：将本地服务注册为集群里的Service
- Forward：将集群服务映射到本地，结合Preview命令可实现开发者之间跨主机使用Localhost地址互访

<!-- tabs:start -->

#### ** Preview命令 **

将本地运行的服务实例注册到集群上。主要用于将本地开发中的服务提供给其他开发者进行联调和预览。

使用`ktctl preview`命令将运行在本地`8080`端口的服务注册到测试集群，命名为`tomcat-v2`。

```bash
$ ktctl preview tomcat-v2 --expose 8080
00:00AM INF KtConnect start at <PID>
... ...
---------------------------------------------------------------
 Now you can access your local service in cluster by name 'tomcat-v2'
---------------------------------------------------------------
```

现在集群里的服务就可以通过`tomcat-v2`名称来访问本地注册的服务实例了，其他开发者也可以在执行`ktctl connect`后，直接通过`tomcat-v2`服务名称来预览该服务的实时情况：

```bash
$ curl http://tomcat-v2:8080
kt-connect local v2
```

#### ** Forward命令 **

将任意IP或集群中的服务映射到本地的指定端口。用于在测试时，使用`localhost`地址便捷的访问集群中的特定IP或服务，典型场景是是访问其他开发者通过Preview命令注册的本地服务。

```text
         ┌─────────────────────────────┐
      forward           |           preview
┌────────┴───────┐      |      ┌───────▼──────┐
│ localhost:8080 │      |      │ local tomcat │
└────────────────┘      |      └──────────────┘
      开发者 B           |           开发者 A
```

例如当一个开发者A运行了前述的Preview命令后，另一个开发者B可以使用`ktctl forward`命令将它映射到自己本地的`6060`端口。

```bash
$ ktctl forward tomcat-v2 6060:8080
00:00AM INF KtConnect start at <PID>
... ...
---------------------------------------------------------------
 Now you can access port 8080 of service 'tomcat-v2' via 'localhost:6060'
---------------------------------------------------------------
```

现在开发者B就可以使用`localhost:6060`地址访问到开发者A本地运行的Tomcat服务了。

当映射的流量源是集群中的服务名时，其效果与`kubectl port-forward`命令相似，只是额外增加了断网自动重连的能力。

<!-- tabs:end -->

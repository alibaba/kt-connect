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

查询Pod和服务地址：

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

使用`connect`命令建立从本地到集群的网络通道：

<!-- tabs:start -->

#### ** MacOS/Linux **

> 在MacOS/Linux环境下默认采用`vpn`模式，该模式采用[sshuttle](https://github.com/sshuttle/sshuttle)工具建立本地与集群的虚拟网络通道，请确保本地具有Python 3.6+运行环境。

使用`connect`命令建立本地到集群的"类VPN"网络隧道，注意该命令需要管理员权限，普通用户需加`sudo`执行：

```bash
$ sudo ktctl connect
00:00AM INF KtConnect start at <PID>
... ...
```

现在本地已经能够直接访问集群资源了，可通过浏览器或`curl`命令来验证：

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

#### ** Windows **

> 在Windows环境下默认采用`socks`模式，该模式本质是在本地创建可访问集群网络的流量代理服务

使用`connect`命令建立本地到集群的Socks代理：

```bash
> ktctl connect                     
00:00AM INF KtConnect start at <PID>
... ...
```

打开一个新的控制台窗口，根据`ktctl connect`命令输出的提示，设置`http_proxy`环境变量。如使用CMD时，命令如下：

```bash
> set http_proxy=socks://127.0.0.1:2223    # 若使用PowerShell，则提示命令应为 $env:http_proxy="socks://127.0.0.1:2223"
```

然后使用`curl`命令来验证：

```bash
> curl http://10.51.0.162:8080    # 在本地直接访问PodIP
kt-connect demo v1

> curl http://172.21.6.39:8080    # 在本地访问ClusterIP
kt-connect demo v1

> curl http://tomcat:8080         # 使用<service>作为域名访问服务
kt-connect demo v1

> curl http://tomcat.default:8080     # 使用<servicename>.<namespace>域名访问服务
kt-connect demo v1

> curl http://tomcat.default.svc.cluster.local:8080    # 使用集群内完整域名访问服务
kt-connect demo v1
```

也可以在浏览器中进行访问，只需将浏览器的代理设置为`socks=127.0.0.1:2223`，详见[Windows支持](zh-cn/guide/windows-support.md)文档。

> 注意 1：对于非管理员用户，域名访问功能不可用，仅支持访问Pod IP和服务的Cluster IP
>
> 注意 2：在**PowerShell**中`curl`与内置命令重名，需用`curl.exe`替代上述命令中的`curl`

<!-- tabs:end -->

## 将集群流量转发到本地

为了验证集群访问本地服务的场景，我们在本地也启动一个Tomcat的容器，并为其创建一个内容不同的首页。

```bash
$ docker run -d --name tomcat -p 8080:8080 tomcat:9
$ docker exec tomcat /bin/bash -c 'mkdir webapps/ROOT; echo "kt-connect local v2" > webapps/ROOT/index.html'
```

KtConnect提供了三种能够让集群访问本地服务的命令，分别用于不同的调试场景。

- Exchange：将集群指定服务的所有流量转向本地
- Mesh：将集群指定服务的部分流量（按Mesh规则）转向本地
- Provide：在集群中创建一个新服务，并将其流量转向本地

<!-- tabs:start -->

#### ** Exchange **

将集群里访问指定服务的所有请求拦截并转发到本地的指定端口，通常用于调试在测试环境调用链上的指定服务。

```text
┌──────────┐     ┌─ ── ── ──     ┌──────────┐
│ ServiceA ├─┬─►x│ ServiceB │ ┌─►│ ServiceC │
└──────────┘ │    ── ── ── ─┘ │  └──────────┘
         exchange             │
             │   ┌──────────┐ │
             └──►│ ServiceB'├─┘
                 └──────────┘
```

由于历史原因，`ktctl exchange`的参数指定的是要替换的目标Deployment名称（而非直接指定Service名称），使用以下命令将先前部署到集群中的`tomcat`服务流量全部转到本地`8080`端口：

```bash
$ ktctl exchange tomcat --expose 8080
00:00AM INF KtConnect start at <PID>
... ...
```

在本地或者集群中访问开头部署到集群的`tomcat`服务，查看输出结果：

> 注意如果未运行`ktctl connect`，只能从集群内访问

```bash
$ curl http://tomcat:8080
kt-connect local v2
```

可以看到，访问集群里`tomcat`服务的请求被路由到了本地的Tomcat实例，现在就可以直接在本地调试这个服务了。

## ** Mesh **

将集群里访问指定服务的部分请求拦截并转发到本地的指定端口。

在默认的`manual`模式下，KtConnect不会自动创建相应的路由规则，Mesh命令运行后，访问该服务的流量将随机访问集群服务和本地实例。您可以自行使用任何服务网格组件（譬如Istio）创建基于`kt-version`标签的路由规则，将特定流量转发到本地。

从`0.2.3`版本开始新增了`auto`模式，该模式不再需要额外的服务网格组件，能够直接实现HTTP请求的自动按需路由。

```text
┌──────────┐     ┌──────────┐    ┌──────────┐
│ ServiceA ├─┬──►│ ServiceB │─┬─►│ ServiceC │
└──────────┘ │   └──────────┘ │  └──────────┘
            mesh              │
             │   ┌──────────┐ │
             └──►│ ServiceB'├─┘
                 └──────────┘
```

以`auto`模式为例。为了便于验证结果，先重置一下集群里Tomcat服务的首页内容。然后通过`ktctl mesh`命令创建代理Pod：

```bash
$ kubectl exec deployment/tomcat -c tomcat -- /bin/bash -c 'mkdir webapps/ROOT; echo "kt-connect demo v1" > webapps/ROOT/index.html'

$ ktctl mesh tomcat --expose 8080 --method auto
00:00AM INF KtConnect start at <PID>
... ...
--------------------------------------------------------------
 Now you can access your service by header 'KT-VERSION: feo3x' 
--------------------------------------------------------------
```

在命令日志的末尾，输出了一个特定的Header值。此时，直接访问集群里的`tomcat`服务，流量将正常进入集群的服务实例：

```bash
$ curl http://tomcat:8080
kt-connect local v2
```

若请求包含Mesh命令输出的Header，则流量将自动被本地的服务实例接收。

```bash
$ curl -H 'KT-VERSION: feo3x' http://tomcat:8080
kt-connect demo v1
```

`ktctl exchange`与`ktctl mesh`命令的最大区别在于，前者会将原应用实例流量全部替换为由本地服务接收，而后者仅将包含指定Header的流量导流到本地，同时保证测试环境正常链路始终可用。

#### ** Provide **

将本地运行的服务实例注册到集群上。与前两种命令不同，`ktctl provide`主要用于将本地开发中的服务提供给其他开发者进行联调和预览。

以下命令会将运行在本地`8080`端口的服务注册到测试集群，命名为`tomcat-preview`。

```bash
$ ktctl provide tomcat-preview --expose 8080
00:00AM INF KtConnect start at <PID>
... ...
```

现在集群里的服务就可以通过`tomcat-preview`名称来访问本地暴露的服务实例了，其他开发者也可以在执行`ktctl connect`后，直接通过`tomcat-preview`服务名称来预览该服务的当前情况：

```bash
$ curl http://tomcat-preview:8080
kt-connect local v2
```

<!-- tabs:end -->

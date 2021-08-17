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

> 在MacOS/Linux环境下默认采用`vpn`模式，该模式采用[sshuttle](https://github.com/sshuttle/sshuttle)工具建立本地与集群的虚拟网络通道，请确保本地具有Python 3.6+运行环境，并通过`pip3 install sshuttle`命令预先安装此工具。

使用`connect`命令建立本地到集群的VPN网络，注意该命令需要管理员权限，普通用户需加`sudo`执行：

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

> 在Windows环境下默认采用`socks`模式

使用`connect`命令建立本地到集群的Socks代理，注意该命令需要当前用户具有管理员权限：

```bash
$ ktctl connect                     
00:00AM INF KtConnect start at <PID>
... ...
```

现在本地已经能够直接访问集群资源了。

由于环境变量和系统代理的修改对已运行的进程无效，在Windows环境中，`ktctl connect`仅对该命令执行之后新创建的进程自动生效。

打开一个新的浏览器或控制台（新窗口，而不是新Tab页），然后在浏览器输入以下地址或使用`curl`命令来验证：

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

> 注意 1：在**PowerShell**中`curl`与内置命令重名，需用`curl.exe`替代上述命令中的`curl`
>
> 注意 2：KtConnect的`socks`模式本质是系统全局代理，但在Windows下并非所有软件都会遵循系统代理配置。譬如基于Spring的Java应用开发可参考[在IDEA中联调](zh-cn/guide/how-to-use-in-idea.md)文档

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

将集群里访问指定服务的部分请求拦截并转发到本地的指定端口。KtConnect不会自动创建相应的路由规则，因此默认情况下访问该服务的流量将随机访问集群服务和本地实例。

您可以自行使用任何Service Mesh工具（譬如Istio）创建基于`version`标签的路由规则，将特定流量转发到本地。

```text
┌──────────┐     ┌──────────┐    ┌──────────┐
│ ServiceA ├─┬──►│ ServiceB │─┬─►│ ServiceC │
└──────────┘ │   └──────────┘ │  └──────────┘
            mesh              │
             │   ┌──────────┐ │
             └──►│ ServiceB'├─┘
                 └──────────┘
```

为了便于验证结果，先重置一下集群里Tomcat服务的首页内容。然后通过`ktctl mesh`命令创建代理Pod：

```bash
$ kubectl exec deployment/tomcat -c tomcat -- /bin/bash -c 'mkdir webapps/ROOT; echo "kt-connect demo v1" > webapps/ROOT/index.html'

$ ktctl mesh tomcat --expose 8080  
00:00AM INF KtConnect start at <PID>
... ...
```

在没有任何额外规则的情况下，访问集群里的`tomcat`服务，流量将随机被路由到本地或集群的服务实例：

```bash
$ curl http://tomcat:8080
kt-connect local v2

$ curl http://tomcat:8080
kt-connect demo v1
```

`ktctl mesh`与`ktctl exchange`命令的最大区别在于，后者会完全替换原有的应用实例，而前者在创建代理Pod后，依然会保留原服务的Pod，代理Pod会动态生成`version`标签，以便用于可以通过Mesh流量规则将特定的流量转发到本地，同时保证测试环境正常链路始终可用。

> 查看更多：[Mesh最佳实践](/zh-cn/guide/mesh)

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

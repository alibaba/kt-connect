# Quick Start Guide

In this chapter, we will deploy a demo app in Kubernetes cluster, use KtConnect to access it from local environment, and redirect request in cluster to a local service.

## Create A Demo App In Cluster

Let's create a simple tomcat deployment, expose it as a service, and put a default html file as index page.

```bash
$ kubectl create deployment tomcat --image=tomcat:9 --port=8080
deployment.apps/tomcat created

$ kubectl expose deployment tomcat --port=8080 --target-port=8080
service/tomcat exposed

$ kubectl exec deployment/tomcat -c tomcat -- /bin/bash -c 'mkdir webapps/ROOT; echo "kt-connect demo v1" > webapps/ROOT/index.html'
```

Fetch information of pod and service created above.

```bash
$ kubectl get pod -o wide --selector app=tomcat
NAME     READY   STATUS    RESTARTS   AGE   IP            ...
tomcat   1/1     Running   0          34s   10.51.0.162   ...

$ kubectl get svc tomcat
NAME     TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)    AGE
tomcat   ClusterIP   172.16.255.111   <none>        8080/TCP   34s
```

Remember the **Pod IP** (`10.51.0.162`) and **Cluster IP** (`172.16.255.111`) of tomcat service, they'll be used in following steps.

## Access App From Local

KtConnect use command `connect` to let all cluster resources accessible from local.

<!-- tabs:start -->

#### ** MacOS/Linux **

> KtConnect run in `vpn` mode in MacOS/Linux by default, this mode use `sshuttle` tool must be installed in order to establish connection between local environment and cluster network.
>
> - MacOS：`brew install sshuttle`
> - Debian/Ubuntu：`apt-get install sshuttle`
> - RedHat/CentOS/Fedora：`dnf install sshuttle`
>
> Please check [sshuttle doc](https://github.com/sshuttle/sshuttle#obtaining-sshuttle) for other operating systems.

The `connect` command require root access to tackle networking configurations, `sudo` should be used for none-root users.

```bash
$ sudo ktctl connect
00:00AM INF KtConnect start at <PID>
... ...
```

Now any resource in the cluster can be directly accessed from local, you could use `curl` or web browser to verify.

```bash
$ curl http://10.51.0.162:8080    # access Pod IP from local
kt-connect demo v1

$ curl http://172.21.6.39:8080    # access Cluster IP from local
kt-connect demo v1

$ curl http://tomcat:8080         # access Service via its name as domain name
kt-connect demo v1

$ curl http://tomcat.default:8080     # access Service via its name and namespace as domain name
kt-connect demo v1

$ curl http://tomcat.default.svc.cluster.local:8080    # access Service via fully qualified domain name
kt-connect demo v1
```

#### ** Windows **

> KtConnect run in `socks` mode in Windows by default.

The `connect` command will create a **Socks Proxy** to tackle all network requests to cluster, Administration permission is required in order to update the resister configuration.

```bash
$ ktctl connect                     
00:00AM INF KtConnect start at <PID>
... ...
```

Now any resource in the cluster can be directly accessed from local.

As global proxy setting and environment variable changes are not applied to application already running in Windows, only new processes created after `ktctl connect` started will be effected.

You could open a new web browser window (not new tab), or use `curl` in a new console to verify.

```bash
$ curl http://10.51.0.162:8080    # access Pod IP from local
kt-connect demo v1

$ curl http://172.21.6.39:8080    # access Cluster IP from local
kt-connect demo v1

$ curl http://tomcat:8080         # access Service via its name as domain name
kt-connect demo v1

$ curl http://tomcat.default:8080     # access Service via its name and namespace as domain name
kt-connect demo v1

$ curl http://tomcat.default.svc.cluster.local:8080    # access Service via fully qualified domain name
kt-connect demo v1
```

> Note 1：In **PowerShell**, the `curl` tool is conflict with a build-in command, please type `curl.exe` instead of `curl`.
>
> Note 2：The `socks` mode use global proxy setting to handle request to cluster resources. however, not all software under Windows follows the system proxy routine, e.g. you will need an [IDE plugin](zh-cn/guide/how-to-use-in-idea.md) when developing Spring based Java application.

<!-- tabs:end -->

## Forward cluster traffic to local

In order to verify the scenarios where the service in cluster access services at local, we also start a Tomcat service locally and create an index page with different content.

```bash
$ docker run -d --name tomcat -p 8080:8080 tomcat:9
$ docker exec tomcat /bin/bash -c 'mkdir webapps/ROOT; echo "kt-connect local v2" > webapps/ROOT/index.html'
```
KtConnect offers 3 commands for accessing local service from cluster in different cases.

- Exchange：redirect all request to specified service in cluster to local
- Mesh：redirect partial request to specified service in cluster (base on mesh rule) to local
- Provide：create a new service in cluster, and redirect any request to this service to local

<!-- tabs:start -->

#### ** Exchange **

Intercept and forward all requests to the specified service in the cluster to specified local port, which is usually used to debug service in the test environment at the middle of an invocation-chain.

```text
┌──────────┐     ┌─ ── ── ──     ┌──────────┐
│ ServiceA ├─┬─►x│ ServiceB │ ┌─►│ ServiceC │
└──────────┘ │    ── ── ── ─┘ │  └──────────┘
         exchange             │
             │   ┌──────────┐ │
             └──►│ ServiceB'├─┘
                 └──────────┘
```

Due to historical reasons, the target Deployment name is used (instead of the Service name) as parameter of `ktctl exchange`. Below command will transfer all the traffic of the `tomcat` service previously deployed in the cluster to `8080` port of developer's laptop.

```bash
$ ktctl exchange tomcat --expose 8080
00:00AM INF KtConnect start at <PID>
... ...
```

Visit the `tomcat` service deployed to the cluster, and view the output result.

> Note: you could also execute below command from any pod in the cluster to verify result, if you didn't run `ktctl connect` command locally.

```bash
$ curl http://tomcat:8080
kt-connect local v2
```

The request to `tomcat` service in the cluster is now routed to the local Tomcat instance, thus you can directly debug this service locally.

## ** Mesh **

Intercept and forward part of requests to the specified service in the cluster to specified local port. Please notice, KtConnect will not automatically create the corresponding routing rules for you, and by default, the traffic accessing the service will randomly access the service in cluster and service at local.

You can use any Service Mesh tool (such as Istio) to create routing rules based on the `version` label of proxy pod to forward specific traffic to the local.

```text
┌──────────┐     ┌──────────┐    ┌──────────┐
│ ServiceA ├─┬──►│ ServiceB │─┬─►│ ServiceC │
└──────────┘ │   └──────────┘ │  └──────────┘
            mesh              │
             │   ┌──────────┐ │
             └──►│ ServiceB'├─┘
                 └──────────┘
```

In order to verify the results, firstly you may need to reset the content of the index page of the Tomcat service in the cluster. Then use `ktctl mesh` command to create a proxy pod:

```bash
$ kubectl exec deployment/tomcat -c tomcat -- /bin/bash -c 'mkdir webapps/ROOT; echo "kt-connect demo v1" > webapps/ROOT/index.html'

$ ktctl mesh tomcat --expose 8080  
00:00AM INF KtConnect start at <PID>
... ...
```

Without any additional rules, the traffic to the `tomcat` service in the cluster will be randomly routed to the local or cluster service instance.

```bash
$ curl http://tomcat:8080
kt-connect local v2

$ curl http://tomcat:8080
kt-connect demo v1
```

The most significant difference between `ktctl mesh` and `ktctl exchange` commands is that the latter will completely replace the original application instance, while the former will still retain the original service pod after the proxy pod is created, and the proxy pod will dynamically generate a `version` label, so that specific traffic can be forwarded to local through mesh rules, while ensuring that the normal traffic in the test environment is not effected.

> Read [Mesh Best Practice](/zh-cn/guide/mesh) doc for more detail

#### ** Provide **

Register a local service instance to cluster. Unlike the previous two commands, `ktctl provide` is mainly used to debug or preview a services under developing.

The following command will register the service running on the port `8080` locally to the cluster as a service named `tomcat-preview`.

```bash
$ ktctl provide tomcat-preview --expose 8080
00:00AM INF KtConnect start at <PID>
... ...
```

Now other services in the cluster can access the locally exposed service instance through the service name of `tomcat-preview`, and other developers can also preview the service directly through service name `tomcat-preview` after executing `ktctl connect` on their laptops.

```bash
$ curl http://tomcat-preview:8080
kt-connect local v2
```

<!-- tabs:end -->

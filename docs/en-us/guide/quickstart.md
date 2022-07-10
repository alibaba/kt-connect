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

KtConnect use command `connect` to let all cluster resources accessible from local (Root/Administrator permission is required).

<!-- tabs:start -->

#### ** MacOS/Linux **

Execute `ktctl connect` command on Mac/Linux with `sudo`:

```bash
$ sudo ktctl connect
00:00AM INF KtConnect start at <PID>
... ...
00:00AM INF ---------------------------------------------------------------
00:00AM INF  All looks good, now you can access to resources in the kubernetes cluster
00:00AM INF ---------------------------------------------------------------
```

#### ** Windows **

On Windows, execute `ktctl connect` command in terminal (If you are not login as Administrator, you should right-click on CMD and PowerShell icon, choose "Run as Administrator" to open a terminal):

```bash
> ktctl connect
00:00AM INF KtConnect start at <PID>
... ...
00:00AM INF ---------------------------------------------------------------
00:00AM INF  All looks good, now you can access to resources in the kubernetes cluster
00:00AM INF ---------------------------------------------------------------
```

<!-- tabs:end -->

Now any resource in the cluster can be directly accessed from local, you could use `curl` or web browser to verify.

> Note：In **Windows PowerShell**, the `curl` tool is conflict with a build-in command, please type `curl.exe` instead of `curl`.

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

<!-- tabs:end -->

## Forward cluster traffic to local

In order to verify the scenarios where the service in cluster access services at local, we also start a Tomcat service locally and create an index page with different content.

```bash
$ docker run -d --name tomcat -p 8080:8080 tomcat:9
$ docker exec tomcat /bin/bash -c 'mkdir webapps/ROOT; echo "kt-connect local v2" > webapps/ROOT/index.html'
```
KtConnect offers 2 commands for accessing local service from cluster in different cases.

- Exchange：redirect all request to specified service in cluster to local
- Mesh：redirect partial request to specified service in cluster (base on mesh rule) to local

<!-- tabs:start -->

#### ** Exchange Command **

Intercept and forward all requests to the specified service in the cluster to specified local port, which is usually used to debug service in the test environment at the middle of an invocation-chain.

```text
┌──────────┐     ┌─ ── ── ──     ┌──────────┐
│ ServiceA ├─┬─►x│ ServiceB │ ┌─►│ ServiceC │
└──────────┘ │    ── ── ── ─┘ │  └──────────┘
         exchange             │
             │   ┌──────────┐ │
             └──►│ ServiceB'├─┘ (Local Instance)
                 └──────────┘
```

Below command will transfer all the traffic of the `tomcat` service previously deployed in the cluster to `8080` port of developer's laptop.

```bash
$ ktctl exchange tomcat --expose 8080
00:00AM INF KtConnect start at <PID>
... ...
---------------------------------------------------------------
 Now all request to service 'tomcat' will be redirected to local
---------------------------------------------------------------
... ...
```

Visit the `tomcat` service deployed to the cluster, and view the output result.

> Note: you could also execute below command from any pod in the cluster to verify result, if you didn't run `ktctl connect` command locally.

```bash
$ curl http://tomcat:8080
kt-connect local v2
```

The request to `tomcat` service in the cluster is now routed to the local Tomcat instance, thus you can directly debug this service locally.

## ** Mesh Command **

Intercept and forward part of requests to the specified service in the cluster to specified local port. Usual used in team cooperation, one developer need to debug a service while do not want to disturb other developers.

```text
┌──────────┐     ┌──────────┐    ┌──────────┐
│ ServiceA ├─┬──►│ ServiceB │─┬─►│ ServiceC │
└──────────┘ │   └──────────┘ │  └──────────┘
            mesh              │
             │   ┌──────────┐ │
             └──►│ ServiceB'├─┘ (Local Instance)
                 └──────────┘
```

Mesh command has 2 execution modes. The default `auto` mode will automatically create corresponding routing rules for you **without** extra service mesh component.

In order to verify the results, firstly let's reset the content of the index page of the Tomcat service in the cluster, then use `ktctl mesh` command to create a traffic rule:

```bash
$ kubectl exec deployment/tomcat -c tomcat -- /bin/bash -c 'mkdir webapps/ROOT; echo "kt-connect demo v1" > webapps/ROOT/index.html'

$ ktctl mesh tomcat --expose 8080  
00:00AM INF KtConnect start at <PID>
... ...
--------------------------------------------------------------
 Now you can access your service by header 'VERSION: feo3x' 
--------------------------------------------------------------
```

At the end of command log, a special HTTP Header is printed. Now if you access the `tomcat` service as normal, the traffic would come into the original pods.

```bash
$ curl http://tomcat:8080
kt-connect demo v1
```

If a request contains the HTTP Header shown by mesh command, the traffic will be redirected to local.

```bash
$ curl -H 'VERSION: feo3x' http://tomcat:8080
kt-connect local v2
```

In actual use, it can be combined with the [ModHeader plugin](https://github.com/bewisse/modheader), so that only the requests made by developers from their browsers will access their local service processes.

The `manual` mode of mesh command provides possibility of more flexible route rule. In this mode, KtConnect will not automatically create the corresponding routing rules for you, the traffic accessing the service will randomly access the service in cluster and service at local. You can use any Service Mesh tool (such as Istio) to create routing rules based on the `version` label of shadow pod to forward specific traffic to the local. Read [Mesh Best Practice](/zh-cn/guide/mesh) doc for more detail

The most significant difference between `ktctl mesh` and `ktctl exchange` commands is that the latter will completely replace the original application instance, while the former will still retain the original service pod after the shadow pod is created, and the router pod will dynamically generate a `version` header (or label), so only specified traffic will be redirected to local, while ensuring that the normal traffic in the test environment is not effected.

<!-- tabs:end -->

## Provide local service to others

In addition to the services that have been deployed to the cluster, during the development process, KtConnect can also be used to quickly "put" a local service to the cluster and turn it into a temporary service for other developers or other services in the cluster to use.

- Preview: Register a local service as a service in the cluster
- Forward: Redirect a local port to a cluster service. Can achieve to access services run on other developer's laptop via localhost while combining with `preview` command

<!-- tabs:start -->

#### ** Preview Command **

Register a local service instance to cluster. Unlike the previous two commands, `ktctl preview` is mainly used to debug or preview a services under developing.

The following command will register the service running on the port `8080` locally to the cluster as a service named `tomcat-v2`.

```bash
$ ktctl preview tomcat-v2 --expose 8080
00:00AM INF KtConnect start at <PID>
... ...
---------------------------------------------------------------
 Now you can access your local service in cluster by name 'tomcat-v2'
---------------------------------------------------------------
... ...
```

Now other services in the cluster can access the locally exposed service instance through the service name of `tomcat-v2`, and other developers can also preview the service directly through service name `tomcat-v2` after executing `ktctl connect` on their laptops.

```bash
$ curl http://tomcat-v2:8080
kt-connect local v2
```

#### ** Forward Command **

Redirect specified local port to any IP or service in the cluster. It is used to easily access a specific IP or service in the cluster using the `localhost` address during testing. The typical scenario is to access local services of other developers registered by `preview` command.

```text
         ┌─────────────────────────────┐
      forward           |           preview
┌────────┴───────┐      |      ┌───────▼──────┐
│ localhost:8080 │      |      │ local tomcat │
└────────────────┘      |      └──────────────┘
    Developer B         |         Developer A
```

For example, after a developer A runs the aforementioned `preview` command, another developer B can use the `ktctl forward` command to map it to its own local `6060` port.

```bash
$ ktctl forward tomcat-v2 6060:8080
00:00AM INF KtConnect start at <PID>
... ...
---------------------------------------------------------------
 Now you can access port 8080 of service 'tomcat-v2' via 'localhost:6060'
---------------------------------------------------------------
```

Now developer B can use the `localhost:6060` address to access the Tomcat service running locally by developer A.

When the forwarded traffic source is a service name in the cluster, the result is similar to the `kubectl port-forward` command, except the additional ability to automatically reconnect when the network is disconnected.

<!-- tabs:end -->

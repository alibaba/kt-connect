使用DNS服务
---

在KtConnect的实现原理中，Shadow容器作为SSH通道用于打通本地与集群之间网络以实现双向的网络访问。 同时Shadow容器也内置了DNS服务，以支持Kubernetes集群的DNS域名解析。 

而在特定的使用场景下，用户的本地开发环境已经使用VPN打通了本地到集群的网络访问（PodIP, ClusterIP），因此只需要使用DNS服务来实现集群的DNS解析能力。 并配合exchange和mesh完成集群到本地的联调场景。

## 部署DNS服务

```
$ kubectl apply -f https://rdc-incubators.oss-cn-beijing.aliyuncs.com/dns/dns.yaml
```

```
$ kubectl get svc kt-connect-dns
NAME             TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)   AGE
kt-connect-dns   ClusterIP   172.21.14.147   <none>        53/UDP    5d22h
```

在本地配置kt-connect-dns服务的PodIP作为DNS Server.

dns.yaml的实际内容如下所示：

```
apiVersion: v1
kind: Service
metadata:
  labels:
    kt-component: dns-server
  name: kt-connect-dns
spec:
  ports:
  - name: dns
    port: 53
    protocol: UDP
    targetPort: 53
  - name: dns-tcp
    port: 53
    protocol: TCP
    targetPort: 53
  selector:
    kt-component: dns-server
  type: ClusterIP
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    kt-component: dns-server
  name: kt-connect-dns
spec:
  replicas: 1
  selector:
    matchLabels:
      kt-component: dns-server
  template:
    metadata:
      labels:
        kt-component: dns-server
    spec:
      containers:
      - image: registry.cn-hangzhou.aliyuncs.com/rdc-incubator/kt-connect-shadow:stable
        name: dns-server
```

## 调试DNS服务

构建DNS服务调试镜像(并推送到仓库)：

```
$ make build-shadow-dlv
```

部署调试版本DNS服务：

```
$ kubectl apply -f https://rdc-incubators.oss-cn-beijing.aliyuncs.com/dns/dns-debug.yaml
```

由于开启了dlv远程调试模式，程序启动后会停留在"API server listening at: [::]:2345"输出的地方，然后等待IDE连接。

查看创建的`kt-connect-dns`服务IP：

```
$ kubectl get svc kt-connect-dns
NAME             TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)   AGE
kt-connect-dns   ClusterIP   172.21.14.147   <none>        53/UDP    5d22h
```

如果本地和集群之间的VPN未建立，用`ktctl connect`打通网络。
通过IDE连接`kt-connect-dns`服务IP的`2345`端口，可以同步查看`kt-connect-dns`服务输出的日志：

```
$ kubectl logs $(kubectl get pod -l kt-component=dns-server -o jsonpath='{.items[0].metadata.name}') dns-server -f
```

使用`nslookup <查询域名> <DNS服务IP>`命令触发查询，例如：

```
$ nslookup tomcat.default.svc.cluster.local 172.21.14.147
```

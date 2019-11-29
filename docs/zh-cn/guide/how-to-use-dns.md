# 使用DNS服务

在KT Connect的实现原理中，Shadow容器作为SSH通道用于打通本地与集群之间网络以实现双向的网络访问。 同时Shadow容器也内置了DNS服务，以支持Kubernetes集群的DNS域名解析。 

而在特定的使用场景下，用户的本地开发环境已经使用VPN打通了本地到集群的网络访问（PodIP, ClusterIP），因此只需要使用DNS服务来实现集群的DNS解析能力。 并配合exchange和mesh完成集群到本地的联调场景。

## 部署DNS服务

```
$ kubectl apply -f https://rdc-incubators.oss-cn-beijing.aliyuncs.com/dns/dns.yaml
```

```
$ kubectl get pods --selector="kt-component=dns-server" -o wide
NAME                              READY   STATUS    RESTARTS   AGE   IP             NODE                        NOMINATED NODE
kt-connect-dns-56dc4597b9-2tb4z   1/1     Running   0          26m   172.23.0.253   cn-beijing.192.168.69.185   <none>
```

在本地配置kt-connect-dns容器的PodIP作为DNS Server.

dns.yaml的实际内容如下所示：

```
apiVersion: extensions/v1beta1
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
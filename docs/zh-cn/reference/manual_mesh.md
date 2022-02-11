基于Ktctl Mesh和服务网格的开发测试示例
---

在本示例中，我们将在集群中部署服务（Tomcat:7）并通过Istio Gateway访问，在确保原有链路可正常访问的情况下通过`ktctl mesh`加入本地联调端点。最后修改Istio路由规则，让只有满足特定规则的流量转发到本地的待调试服务（Tomcat:8）。

### 准备实例应用程序

> 前置条件，当前Kubernetes集群已经部署Istio组件

创建命名空间，并启用Istio自动注册：

```bash
$ kubectl create namespace mesh-demo
$ kubectl label namespace mesh-demo istio-injection=enabled
```

在集群中准备示例应用：

``` yaml 
# tomcat7-deploy.yaml
apiVersion: v1
kind: Service
metadata:
  name: tomcat
spec:
  ports:
    - port: 8080
      protocol: TCP
      targetPort: 8080
  selector:
    run: tomcat
  sessionAffinity: None
  type: ClusterIP
status:
  loadBalancer: {}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    run: tomcat
    version: v1
  name: tomcat
spec:
  selector:
    matchLabels:
      run: tomcat
      version: v1
  template:
    metadata:
      labels:
        run: tomcat
        version: v1
    spec:
      containers:
        - image: 'tomcat:7'
          name: tomcat
          ports:
            - containerPort: 8080
              protocol: TCP
```

部署应用：

```bash
$ kubectl -n mesh-demo apply -f tomcat7-deploy.yaml
service/tomcat created
deployment.apps/tomcat created
```

### 通过服务网格访问服务

创建默认的Istio路由规则：

```yaml
# tomcat7-istio.yaml
apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: tomcat-gateway
spec:
  selector:
    istio: ingressgateway
  servers:
  - hosts:
    - 'tomcat.mesh.com'
    port:
      name: http
      number: 80
      protocol: HTTP
---
apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: tomcat
spec:
  host: tomcat
  subsets:
  - name: v1
    labels:
      version: v1
---
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: tomcat
spec:
  gateways:
  - tomcat-gateway  # 绑定gateway
  hosts:
  - tomcat.mesh.com
  - tomcat
  http:
  - route:
    - destination:
        host: tomcat
        subset: v1
```

部署服务网格定义：

```bash
$ kubectl -n mesh-demo apply -f tomcat7-istio.yaml
gateway.networking.istio.io/tomcat-gateway created
destinationrule.networking.istio.io/tomcat created
virtualservice.networking.istio.io/tomcat created
```

获取Istio的访问入口

```bash
$ export INGRESS_HOST=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
$ export INGRESS_PORT=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.spec.ports[?(@.name=="http")].port}')
```

这里在本地Hosts中添加自定义DNS：

```text
<INGRESS_HOST> tomcat.mesh.com
```

通过域名`http://tomcat.mesh.com`访问实例应用:

![](../../media/guide/demo-tomcat-7.png)

### Mesh添加本地访问端点

在本地使用tomcat:8容器，并监听本地的8080端口：

```bash
$ docker run -itd -p 8080:8080 tomcat:8
```

添加本地联调端点：

```bash
$ ktctl -n mesh-demo mesh tomcat --expose 8080 --mode manual
00:00AM INF KtConnect start at <PID>
... ...
--------------------------------------------------------------
 Now you can update Istio rule by label 'version: ngzlj' 
--------------------------------------------------------------
```

如上所示，这里部署了一个本地端点，并且版本号为`ngzlj`。此时如果访问`http://tomcat.mesh.com`能正常访问tomcat7。

### 定义本地端点访问规则

修改路径规则，并确保当使用Firefox浏览器访问时，流量转移到本地运行的Tomcat:8，如下所示

```yaml
apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: tomcat
spec:
  host: tomcat
  subsets:
  - name: v1
    labels:
      version: v1
  - name: ngzlj  # 添加本地端点版本
    labels:
      version: ngzlj
---
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: tomcat
spec:
  gateways:
  - tomcat-gateway
  hosts:
  - tomcat.mesh.com
  - tomcat
  http:
  - match:  # 定义路由规则
    - headers: 
        user-agent:  # 匹配请求的user-agent
          exact: Mozilla/5.0 (Macintosh; Intel Mac OS X 10.14; rv:67.0) Gecko/20100101 Firefox/67.0
    route:
    - destination:
        host: tomcat
        subset: ngzlj
  - route:
    - destination:
        host: tomcat
        subset: v1
```

此时，如果通过Firefox浏览器访问服务，则可以访问到本地Tomcat:8实例:

![](../../media/guide/demo-tomcat-8.png)

而通过非Firefox浏览器访问应用则能正常访问到原有的Tomcat:7应用。

![](../../media/guide/demo-tomcat-7.png)

备注：user-agent的可以通过Firefox的浏览器开发工具查看，如下所示：

![](../../media/guide/demo-user-agent.png)
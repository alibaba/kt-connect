Dashboard
====

> Dashboard需配合0.0.4+版本的客户端使用

KT Connect 提供了一个可视化控制台用于查看和管理测试环境的整体状态。如下所示：

![](../../_media/guide/kt-dashboard.png)

目前Dashboard提供的主要功能包括：

* Cli连接情况统计：包括Connect,Exchange,Mesh的连接数
* 本地联调拓扑结构：以服务维度展示联调拓扑结构
* 组件详情： 查看连接端点的详细信息，包括调用路径以及客户端信息

## 如何安装

设置RBAC权限，以使Dashboard组件能够获取和监听Kubernetes集群的资源变化；

```
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: ktadmin
rules:
- apiGroups: [""]
  resources:
  - namespaces
  - nodes
  - nodes/proxy
  - services
  - endpoints
  - pods
  verbs: ["get", "list", "watch"]
- apiGroups:
  - extensions
  resources:
  - ingresses
  verbs: ["get", "list", "watch"]
- nonResourceURLs: ["/metrics"]
  verbs: ["get"]
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: ktadmin
  namespace: default
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: ktadmin
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: ktadmin
subjects:
- kind: ServiceAccount
  name: ktadmin
  namespace: default
```

创建服务和Deployment实例:

```
apiVersion: v1
kind: Service
metadata:
  name: kt-dashboard
spec:
  ports:
  - port: 80
    targetPort: 80
  selector:
    app: kt-dashboard
  type: ClusterIP
---
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  labels:
    app: kt-dashboard
  name: kt-dashboard
spec:
  replicas: 1
  selector:
    matchLabels:
      app: kt-dashboard
  template:
    metadata:
      labels:
        app: kt-dashboard
    spec:
      serviceAccount: ktadmin
      containers:
      - image: registry.cn-hangzhou.aliyuncs.com/rdc-incubator/kt-connect-dashboard:stable
        imagePullPolicy: Always
        name: dashboard
        ports:
        - containerPort: 80
      - image: registry.cn-hangzhou.aliyuncs.com/rdc-incubator/kt-connect-server:stable
        imagePullPolicy: Always
        name: controller
        ports:
        - containerPort: 8000
```

将以上内容分别保存到`rbac-setup.yaml`和`dashboard.yaml`。 并通过kubectl部署：

```
kubectl apply -f rbac-setup.yaml
kubectl apply -f dashboard.yaml
```

或者直接使用一下命令: 

```
kubectl apply -f https://rdc-incubators.oss-cn-beijing.aliyuncs.com/dashboard/stable/rbac.yaml
kubectl apply -f https://rdc-incubators.oss-cn-beijing.aliyuncs.com/dashboard/stable/dashboard.yaml
```


容器正常启动后，通过port-forward在本地访问：

```
$ kubectl port-forward deployments/kt-dashboard 8000:80   
Forwarding from 127.0.0.1:8000 -> 80
Forwarding from [::1]:8000 -> 80
```

打开浏览器http://127.0.0.1:8000
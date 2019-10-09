Dashboard
====

> Cli should upgrade to 0.0.4+

KT Connect Dashboard support visualization your test environment status, as follow:

![](../_media/guide/kt-dashboard.png)

Current support features as follow:

* Connection statistical： how many cli connection to kubernetes cluster.
* Topology View：the topo of the service.
* Compnent Info: You can view the endpoint detil information, eq. the client local address.

## How to install

You can install `KT Connect Dashboard` as follow

Inorder to let Dashboard can watch the resource change, you should Setup RBAC first:

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

And Create new Service And Deployment:

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

You can save below content to `rbac-setup.yaml` and `dashboard.yaml`，and deploy to Kubernetes cluster:

```
kubectl apply -f rbac-setup.yaml
kubectl apply -f dashboard.yaml
```

Or use shell as follow: 

```
kubectl apply -f https://rdc-incubators.oss-cn-beijing.aliyuncs.com/dashboard/stable/rbac.yaml
kubectl apply -f https://rdc-incubators.oss-cn-beijing.aliyuncs.com/dashboard/stable/dashboard.yaml
```

After all components up, you can view in local:

```
$ kubectl port-forward deployments/kt-dashboard 8000:80   
Forwarding from 127.0.0.1:8000 -> 80
Forwarding from [::1]:8000 -> 80
```

And then open http://127.0.0.1:8000
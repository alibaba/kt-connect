## Command: ktctl run

将本地服务暴露到Kubernetes集群

### 示例

```
ktctl run localservice --port 8080 --expose
```

### 参数

```
--port value  The port that exposes (default: 0)
--expose      If true, a public, external service is created
```

### 从父命令集成的参数

```
--namespace value, -n value   (default: "default")
--kubeconfig value, -c value  (default: "/Users/yunlong/.kube/config")
--image value, -i value       Custom proxy image (default: "registry.cn-hangzhou.aliyuncs.com/rdc-incubator/kt-connect-shadow:stable")
--debug, -d                   debug mode
--label value, -l value       Extra labels on proxy pod e.g. 'label1=val1,label2=val2'
--help, -h                    show help
--version, -v                 print the version
```

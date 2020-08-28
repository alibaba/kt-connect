## Command: ktctl mesh

将本地服务混合到集群中

### 示例

```
ktctl --debug --namespace=default mesh tomcat --expose 8080
```

### 参数

```
--expose value, -e value   Ports to expose separate by comma, in [port] or [remote:local] format, e.g. 7001,80:8080
--version-label value      Specify the version of mesh service, e.g. '0.0.1'
```

### 从父命令集成的参数

```
--namespace value, -n value   (default: "default")
--kubeconfig value, -c value  (default: "/Users/yunlong/.kube/config")
--image value, -i value       Custom proxy image (default: "registry.cn-hangzhou.aliyuncs.com/rdc-incubator/kt-connect-shadow:stable")
--debug, -d                   debug mode
--label value, -l value       Extra labels on proxy pod e.g. 'label1=val1,label2=val2'
--help, -h                    Show help
--version, -v                 Print the version
```

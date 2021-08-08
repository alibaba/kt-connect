## Command: ktctl exchange

使用本地服务替换集群中的Deployment实例

### 示例

```
ktctl --debug --namespace=default exchange tomcat --expose 8080
```

### 常用参数

```
--expose value  指定要暴露的一个或多个端口，逗号分隔，格式为`port`或`local:remote`，例如：7001,8080:80
```

### 从父命令集成的参数

```
--namespace value, -n value   目标Namespace名称 (默认值：default)
--kubeconfig value, -c value  使用的Kubernetes集群配置文件 (默认值：~/.kube/config，若存在KUBECONFIG变量则从该变量读取)
--image value, -i value       指定使用的代理镜像 (默认值：registry.cn-hangzhou.aliyuncs.com/rdc-incubator/kt-connect-shadow:<当前版本>)
--debug, -d                   开启调试日志
--label value, -l value       为代理Pod增加额外标签，例如 'label1=val1,label2=val2'
--forceUpdate                 创建代理Pod时强制更新最新镜像版本
```

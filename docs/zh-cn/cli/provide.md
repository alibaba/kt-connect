## Command: ktctl provide

将本地服务添加到Kubernetes集群

### 示例

```
ktctl provide localservice --expose 8080
```

### 常用参数

```
--expose value  指定本地服务监听的端口
--external      创建`LoadBalancer`类型的Service（生成可暴露到集群外的服务地址）
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

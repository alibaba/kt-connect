## 命令: ktctl connect

从本地连接到集群

### 示例

```
ktctl --debug --namespace=default connect --method=socks5
```

### 常用参数

```
--method value         与集群建立虚拟连接的方式，可选值有 'vpn'（仅Linux/Mac）、'tun'（仅Linux）、'socks' 和 'socks5'
--shareShadow          与其他开发者共用代理Pod
--localDomain value    指定本地的域名尾缀（默认值：无）
--clusterDomain value  指定集群的域名尾缀（默认值：cluster.local）
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

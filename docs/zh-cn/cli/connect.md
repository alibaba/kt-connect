## 命令: ktctl connect

从本地连接到集群

### 示例

```
ktctl --debug --namespace=default connect --method=socks5
```

### 常用参数

```
--method           与集群建立虚拟连接的方式，可选值有 'vpn'（仅Linux/Mac）、'tun'（仅Linux）、'socks' 和 'socks5'
--shareShadow      与其他开发者共用代理Pod
--sshPort          指定用于本地SSH映射的端口（默认值：2222）
--dump2hosts       使用本地Hosts文件访问Service域名，并指定需要访问的Namespace列表，用逗号分隔
--clusterDomain    指定集群的域名尾缀（默认值：cluster.local）
--disablePodIp     （仅用于vpn模式）禁用Pod IP访问
--excludeIps       （仅用于vpn模式）将指定IP段指定为非集群网段，用逗号分隔
--cidr             （仅用于vpn模式）增加Pod IP段范围，用逗号分隔
--disableDNS       （仅用于vpn和tun模式）禁用DNS代理
--tunName          （仅用于tun模式）指定本地tun设备名称（默认值：tun0）
--tunCidr          （仅用于tun模式）指定本地tun设备网段（默认值：10.1.1.0/30）
--proxyAddr        （仅用于socks5模式）指定Socks5代理监听的地址（默认值：127.0.0.1）
--proxyPort        （仅用于socks5模式）指定Socks5代理监听的端口（默认值：2223）
--jvmrc            （仅用于Windows系统）生成IDEA插件使用的jvmrc文件
--setupGlobalProxy （仅用于Windows系统）自动配置系统全局代理
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

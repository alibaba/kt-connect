## 命令: ktctl connect

从本地连接到集群

### 示例

```
ktctl --debug --namespace=default connect --method=socks5
```

### 参数

```
--method value  Connect method 'vpn', 'socks' or 'socks5' (default: "vpn")
--proxy value   when should method socks5, you can choice which port to proxy, default 2223 (default: 2223)
--port value    Local SSH Proxy port (default: 2222)
--disableDNS    Disable Cluster DNS
--cidr value    Custom CIDR, e.g. '172.2.0.0/16'
--dump2hosts    Auto write service to local hosts file (客户端版本0.0.10+)
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

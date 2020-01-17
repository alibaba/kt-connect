## Command: ktctl connect

Connection to kubernetes cluster

### Usage

```
ktctl --debug --namespace=default connect --method=socks5
```

### Options

```
--method value  Connect method 'vpn' or 'socks5' (default: "vpn")
--proxy value   when should method socks5, you can choice which port to proxy, default 2223 (default: 2223)
--port value    Local SSH Proxy port (default: 2222)
--disableDNS    Disable Cluster DNS
--cidr value    Custom CIDR eq '172.2.0.0/16'
```

### Global Options

```
--namespace value, -n value   (default: "default")
--kubeconfig value, -c value  (default: "/Users/yunlong/.kube/config")
--image value, -i value       Custom proxy image (default: "registry.cn-hangzhou.aliyuncs.com/rdc-incubator/kt-connect-shadow:stable")
--debug, -d                   debug mode
--label value, -l value       Extra labels on proxy pod e.g. 'label1=val1,label2=val2'
--help, -h                    show help
--version, -v                 print the version
```

## Command: ktctl dashboard

快速安装和使用kt dashboard

### 示例

```
# 安装或者更新KT Dashboard
ktctl dashboard init

# 在本地浏览器打开KT Dashboard
ktctl dashboard open
```

### 参数

```
COMMANDS:
     init  install/update dashboard to cluster
     open  open dashboard

OPTIONS:
   --help, -h  show help
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

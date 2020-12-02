## Command: ktctl dashboard

Install or upgrade dashboard 

### Usage

```
# Install or Upgrade Dashboard
ktctl dashboard init

# Open KT Dashboard
ktctl dashboard open
```

### Options

```
COMMANDS:
     init  install/update dashboard to cluster
     open  open dashboard

OPTIONS:
   --help, -h  show help
```

### Global Options

```
--namespace value, -n value   (default: "default")
--kubeconfig value, -c value  (default: env from KUBECONFIG)
--image value, -i value       Custom proxy image (default: "registry.cn-hangzhou.aliyuncs.com/rdc-incubator/kt-connect-shadow:stable")
--debug, -d                   debug mode
--label value, -l value       Extra labels on proxy pod e.g. 'label1=val1,label2=val2'
--help, -h                    show help
--version, -v                 print the version
```

## Command: ktctl exchange

Exchange a running deployment to local

### Usage

```
ktctl --debug --namespace=default exchange tomcat --expose 8080
```

### Options

```
--expose value  expose port
```

### Global Options

```
--namespace value, -n value   (default: "default")
--kubeconfig value, -c value  (default: env from KUBECONFIG)
--image value, -i value       Custom shadow image (default: "registry.cn-hangzhou.aliyuncs.com/rdc-incubator/kt-connect-shadow:stable")
--debug, -d                   debug mode
--label value, -l value       Extra labels on proxy pod e.g. 'label1=val1,label2=val2'
--help, -h                    show help
--version, -v                 print the version
```

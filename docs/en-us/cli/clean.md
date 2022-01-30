## Command: ktctl clean

Delete unavailing shadow pods from kubernetes cluster

### Usage

```
ktctl clean
```

### Options

```
--dryRun                  Only print name of deployments to be deleted
--thresholdInMinus value  Length of allowed disconnection time before a unavailing shadow pod be deleted (default: 30)
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

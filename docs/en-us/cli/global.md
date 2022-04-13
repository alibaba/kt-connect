Global Options
---

`ktctl` command options are divided into "global parameters" and "subcommand parameters". The global parameters should be directly followed by `ktctl`, and the subcommand parameters should be placed after the specific subcommand. For example:

```bash
$ ktctl --namespace demonstration connect --includeIps 10.1.0.0/16,10.2.0.0/16
        | <-- Global options --> |        | <----- Subcommand options -----> |
```

Available options:

```text
--namespace value, -n value   Specify target namespace (otherwise follow kubeconfig current context)
--kubeconfig value, -c value  Specify path of KubeConfig file (default: "/Users/flin/.kube/config")
--image value, -i value       Customize shadow image (default: "registry.cn-hangzhou.aliyuncs.com/rdc-incubator/kt-connect-shadow:vdev")
--imagePullSecret value       Custom image pull secret
--serviceAccount value        Specify ServiceAccount name for shadow pod (default: "default")
--nodeSelector value          Specify location of shadow and route pod by node label, e.g. 'disk=ssd,region=hangzhou'
--debug, -d                   Print debug log
--withLabel value, -l value   Extra labels on proxy pod e.g. 'label1=val1,label2=val2'
--withAnnotation value        Extra annotation on proxy pod e.g. 'annotation1=val1,annotation2=val2'
--portForwardTimeout value    Seconds to wait before port-forward connection timeout (default: 10)
--podCreationTimeout value    Seconds to wait before shadow or router pod creation timeout (default: 60)
--useShadowDeployment         Deploy shadow container as deployment
--useLocalTime                Use local time (instead of cluster time) for resource heartbeat timestamp
--forceUpdate, -f             Always update shadow image
--context value               Specify current context of kubeconfig
--podQuota value              Specify resource limit for shadow and router pod, e.g. '0.5c,512m'
--help, -h                    show help
--version, -v                 print the version
```

Key options explanation:

- `--namespace` actually specifies which Namespace to run Shadow Pod in.
  For the `connect`, `preview` commands, it will affect the access method of the service, that is, you can directly access the service in the same Namespace as the Shadow Pod through `<ServiceName>`, while accessing other Namespace services must use `<ServiceName>.<Namespace>` as the domain name.
  For `exchange`, `mesh` commands, you must specify the same Namespace as the target service to be replaced.
- `--podQuota` use letter `c` for CPU quota (number of cores), use letter `k`/`m`/`g` for memory quota (amount of "KB"/"MB"/"GB")

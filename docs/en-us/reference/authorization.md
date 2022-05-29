Cluster permissions
---

The `ktctl` command-line tool is required to have operation permissions on resource types such as Pod, Service, and ConfigMap in the process of establishing network channels to a Kubernetes cluster.

If you encounter a permission-related error when using `ktctl`, such as "User \"xxx\" cannot list resource \"yyy\" in API group \"zzz\"", it means the Role or ClusterRole you are currently using doesn't meet the permission requirement for KT.

Below are the minimum RBAC permission configuration required for typical feature sets of KT:

- Minimal permissions, only supports `connect` command running with default parameters: [connect-only-mini.yaml](https://github.com/alibaba/kt-connect/blob/master/docs/deploy/rbac/connect-only-mini.yaml)
- Support full functionality using `connect` command: [connect-only-full.yaml](https://github.com/alibaba/kt-connect/blob/master/docs/deploy/rbac/connect-only-full.yaml)
- Supports all commands with default parameters: [all-commands-mini.yaml](https://github.com/alibaba/kt-connect/blob/master/docs/deploy/rbac/all-commands-mini.yaml)
- Supports full functionality of all commands: [all-commands-full.yaml](https://github.com/alibaba/kt-connect/blob/master/docs/deploy/rbac/all-commands-full.yaml)

For enterprises that have permission control requirements and cannot directly distribute kubeconfig file to developers, you can also use KT's [customized compiled version](en-us/reference/customize.md) to build kubeconfig configuration with corresponding cluster operation permissions into ` ktctl` binary.

Frequency Asked Questions
---

#### Q：For multiple test clusters, how to specify the target of `ktctl` ?

A：`ktctl` will access cluster according to local configuration of `kubectl` tool, which usually lay on `~/.kube/config`.

#### Q: What is the minimal RBAC permission required by `ktctl` client ?

Please check out the [sample](https://github.com/alibaba/kt-connect/blob/feature/minimum-permissions/docs/deploy/rbac/clusterrole.yaml). 

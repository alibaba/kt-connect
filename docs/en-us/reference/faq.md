Frequency Asked Questions
---

#### Q: I have several kubernetes clusters, how to specify the target of `ktctl` ?

A: `ktctl` will access cluster according to local configuration of `kubectl` tool, which usually lay on `~/.kube/config`.

#### Q: What is the minimal RBAC permission required by `ktctl` client ?

A: Please check out this [cluster role yaml](https://github.com/alibaba/kt-connect/blob/feature/minimum-permissions/docs/deploy/rbac/clusterrole.yaml). 

#### Q: Encounter error of "too many open files" under MacOS/Linux ?

A: This is caused by the insufficient upper limit of the number of system file handles. For solutions, please refer to: [MacOS](https://www.jianshu.com/p/d6f7d1557f20) / [Linux](https://zhuanlan.zhihu.com/p/75897823)

#### Q: After executing `exchange` or `mesh` command, a 503 error is reported when accessing the target Service, and the returned Header contains "server: envoy" ?

A: The default `selector` mode of Exchange and the default `auto` mode of Mesh are not compatible with the Istio service mesh. If Istio components are used, please use the `scale` mode of Exchange and the `manual` mode of Mesh.
If the above error still exists after the switch, please check why the `VirtualService` and `DestinationRule` rules on the service cannot select the Shadow Pod created by KT.

#### Q: Encounter error of "unable to do port forwarding: socat not found" when executing `ktctl` command ?

A: The port mapping function of `Ktctl` depends on the `socat` tool on the cluster host, please pre-install it on each node of the cluster (Debian/Ubuntu distribution installation command: `apt-get install socat`, CentOS/RedHat distribution installation command: `yum install socat`)

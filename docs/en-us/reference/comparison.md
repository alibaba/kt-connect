Comparison of related tools
---

`kt-connect` is a cloud-native R&D auxiliary tool derived from the internal technical practice of cloud effects. The problems it hopes to solve are similar to the network tool `telepresence`, but due to differences in technical implementation methods, the two are used in different scenarios. The abilities below are also unique.

From the basic mode of operation. `kt-connect` is an open source internal tool, all functions are completely free, and it supports enterprise [quick customization](en-us/reference/customize.md). `telepresence` is an open source commercial software, some functions can only be used after a subscription fee. From the perspective of the open source customization protocol, the GPL v3 protocol adopted by `kt-connect` does not support the use of the project source code for commercial software (see the [FAQ](en-us/reference/faq.md) document for details), while the source code of telepresence does not have this restriction.

| | kt-connect | telepresence |
| --- | --- | --- |
| Main Developer | Alibaba Cloud Cloud Effect | Ambassador Labs |
| Development Language | Golang | Golang |
| Open Source License | GPL v3 | Apache v2.0 |
| Paid or not | Completely free | Paid features available |
| Customization | Configuration customization or modification of source code | Modification of source code |

In terms of functions and implementation methods, the core capabilities provided by the two are similar, that is, to solve the problems of two-way communication and traffic forwarding of cloud-native cluster networks. On top of this, `telepresence` can provide some Istio-like network governance capabilities and corresponding platform services (requires paid subscription) due to the Sidecar mechanism, but its network traffic rules are incompatible with Istio Agent.

Under normal circumstances, `kt-connect` will completely clear all resources created in the cluster when the client exits, and the resources are less intrusive. `telepresence` will still leave the `ambassador` Namespace and several Services and Pods in the cluster after the client exits, which requires developers to clean up afterward. On the uplink (from the local to the cluster), `kt-connect` encapsulates data based on the `socks5` protocol, while `telepresence` uses the `grpc` protocol and supports IPv6 addresses. In theory, the communication efficiency of the latter is slightly higher ( not tested). On the downstream link (from the cluster to the local), `kt-connect` forwards the traffic at the Service layer, there are no other side effects when establishing the traffic channel, and the connection speed is faster. `telepresence` forwards traffic based on Sidecar at the Pod layer. When a traffic channel is established, all Pods corresponding to the service will be restarted. The duration of the connection establishment is related to the time it takes for the Pod to restart to ready, and there is a slight performance loss due to Sidecar forwarding. Its theoretical communication efficiency is lower.

In terms of resource cost, `kt-connect` creates an independent proxy Pod for each user's uplink by default (which can be configured to be shared), and all uplink traffic of `telepresence` shares a proxy Pod, which consumes less resources. But for downstream traffic, `kt-connect` creates a reverse proxy pod for each service that needs to be redirected, and `telepresence` adds a sidecar to each pod, which consumes more resources.

| | kt-connect | telepresence |
| --- | --- | --- |
| Cluster resource residency | None | One Namespace and several Services and Pods |
| Uplink Side Effects | None | None |
| proxy protocol | socks5 | grpc |
| IP v4 address | Support | Support |
| IP v6 address | Not supported | Supported |
| Downlink Side Effects | None | All Pods Restart |
| Only part of the flow is returned | Supported | Supported |
| Traffic upstream resource cost | High, each user uses an independent proxy Pod | Low, all traffic reuses one proxy Pod |
| Traffic downlink resource cost | Low, one reverse proxy Pod per service | High, one Sidecar container added to each Pod |
| The relationship between upstream and downstream traffic | Can be used independently | You must first establish an upstream channel before continuing to create a downstream channel |
| Expose local services | Support | Support, login required |
| Compatible with IstioAgent | Supported | Not Supported |

In terms of use, `kt-connect` has made many detailed optimizations based on the problems encountered in the use of the cloud effect itself. For example, when supporting local access to cluster services, the pure service name without the `namespace` suffix is ​​used as the domain name (`telepresence` only supports short domain names in the form of `service-name.namespace`), supports full intranet environment cluster (custom proxy Pod image), supports global default configuration items (ktctl config command) and so on. If you need a small and practical Kubernetes cloud-native R&D network interworking tool, `kt-connect` will be a good choice.

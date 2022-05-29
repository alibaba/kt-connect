Changelog
---

### 0.3.5

> Release time: 2022-05-30

- Add new command `config` to support global default parameter value
- `exchange`/`mesh`/`preview` commands support skip local port checking (thanks to @[wuxs](https://github.com/wuxs))
- Add a check on the result of local route table setting in case of half-done
- `connect` command remove the permission requirement of listing cluster namespaces
- Support custom local DNS proxy addresses and order
- Support embedded kubeconfig configuration into `ktctl` binary
- The local configuration directory was renamed from `.ktctl` to `.kt`
- Fixed an issue that modification of hosts file may affect access to intranet domain names (thanks to @[cryice](https://github.com/cryice))
- Fixed a DNS order issue when coexisting with OpenVPN under Windows
- Fixed an issue of local routes were not removed correctly in some Windows environments

### 0.3.4

> Release time: 2022-05-04

- All of `connect`/`exchange`/`mesh`/`preview` commands support auto reconnection after network recover
- Record background task log in debug mode as individual log file
- Fixed an DNS resolving issue introduced in version 0.3.3

### 0.3.3

> Release time: 2022-04-27

- Support put global parameters and subcommand parameters in any order
- `mesh` command now route requests with unknown header values to the default environment, instead of show "404" error
- `exchange` and `mesh` commands support service with target ports defined by name
- `clean` command supports sweeping residual local routing table records
- Show currently connected kubernetes cluster name and the configured context name at startup
- Try to listen to a random port to work around the problem of abnormal port checking logic in some environments
- Fixed an issue of the Router Pod was not correctly deleted when the `mesh` command exited
- Fixed an issue of the Service created by the `preview` command misusing local port number of the `--expose` parameter
- Fixed an issue of the unexpired resources of the cluster were cleaned up due to the inconsistency of local time between developers
- Fixed an issue of the Port Forward failed due to the overlap between the Cluster IP segment and the API Server address
- Fixed an issue of the proxy DNS resolve CName record incorrectly

### 0.3.2

> Release time: 2022-03-28

- Add new command `recover` to immediately restore the traffic changed by `exchange` or `mesh` command of specified service
- `connect` command now automatically clean up expired resources in the cluster before each run, without manually executing the `clean` command
- `connect` command add `--useShadowDeployment` parameter to support deploying shadow containers as Deployment
- `connect` command add `--podQuota` parameter to support configuring the resource limitation of shadow pod and router pod
- `connect` command routing rule no longer dependent on the node's PodCIDR configuration
- `connect` command now monitoring Service changes when using `hosts` dns mode
- Show owner's username when the target service of `exchange`/`mesh` command is occupied
- `manual` mode of the `mesh` command now using Service name as the target parameter
- Fix an issue that the routing setting does not take effect in some cases in Windows environment (thanks to @[dominicqi](https://github.com/dominicqi))
- Fix an issue that CPU and memory usage often soared in Windows environment
- Fix error message when execute `ktctl` command without sub-command

### 0.3.1

> Release time: 2022-02-20

- Support auto merge multiple kubeconfig files
- Support change local dns cache timeout
- Fix an issue of `connect` command `localDNS` mode failed to resolve cluster domain
- Fix an issue of `exchange` command continually printing error even connect already recovered
- Fix an issue cause stuntman service of `mesh` command selecting router pod in `auto` mode
- Fix an issue of `clean` command failed to delete resources left by `exchange` command
- Shorten resource heartbeat interval and lock timeout to speed up expired resource recycling
- Add `ports` field to shadow pods and router pods

### 0.3.0

> Release time: 2022-02-13

- `connect` command supports `tun2socks` mode
- `connect` command supports simultaneous resolution of cluster service domain names and local intranet/extranet domain names
- `connect` command supports access to Headless Service on all systems
- `exchange` command defaults to `selector` mode
- `mesh` command defaults to `auto` mode
- `exchange` and `mesh` uniformly use Service as the target
- Deprecated `dashboard` command
- Deprecated `kubectl` plugin
- Added target port check for `exchange`/`mesh` commands
- Fixed command line parameter validity check

### 0.2.5

> Release time: 2021-12-30

- Optimize `--expose` parameter of `provide` command to support multiple ports and port mapping
- Optimized the cleaning mark method of `clean` command to solve the problem of incomplete resource cleaning
- Optimized calculation logic of the pod IP range of the `connect` command to avoid routing effects on irrelevant IP segments
- Add `--nodeSelector` parameter to support specifying shadow pod to specified node (issue-185)
- Fixed an issue of after auto mesh, service redeployment may cause mesh failure
- Fixed an issue of error throw out when the `clean` command cleans resources whose status is already `Terminating`
- Fixed an issue of the `connect` command reported an error when the local KubeConfig did not have the global pod permission
- Fixed an issue of the `connect` command would have residual resources in some cases of abnormal exit

### 0.2.4

> Release time: 2021-12-23

- Support `auto` mode of `mesh` command to specify access target using service name
- Added `switch` mode for `exchange` command, no longer need to wait for pod resume when exiting
- Remove the default password of shadow pod and use temporary private key to improve security
- Fixed an issue of Namespace must be specified when running `ktctl` when the local kube config file is not configured with namespace
- Fixed an issue of Shadow Pod was not cleaned up probabilistically when using Ctrl+C to interrupt `exchange` to exit the wait
- Fixed an issue of the program could not exit automatically when the startup of sshuttle failed

### 0.2.3

> Release time: 2021-12-09

* `mesh` command supports `auto` mode without istio dependency
* During the process of `exchange` command exit, support using Ctrl+C to interrupt waiting
* `--dump2hosts` parameter of `connect` command supports non-socks mode
* Normalize error log output to display error messages as detailed as possible

### 0.2.2

> Release time: 2021-11-12

* `exchange` command waits for the original service to be fully restored before exiting the shadow pod (issue-257)
* `connect` command adds the `--excludeIps` parameter to exclude the specified non-cluster IP segment
* `connect` command adds the `--proxyAddr` parameter, which is used to specify the IP address that the Socks5 proxy listens to
* `exchange`/`mesh`/`provide` command adds a check to see if there is a service listening on the local port
* Fix an issue of the `--cidr` parameter of the `connect` command specified multiple IP segments incorrectly
* Fix an issue of `exchange` connection automatically disconnected when the local service restarts or the response times out

### 0.2.1

> Release time: 2021-11-07

* Use the namespace of the current context of kubeconfig by default (issue #102)
* Fix `connect` command bug when using shared shadow pod (issue #260)
* Added `--context` global parameter to support switching context in kubeconfig (issue #261)

### 0.2.0

> Release time: 2021-10-17

* Kubernetes minimum compatible version increased to `1.16`
* Use shadow pod instead of shadow deployment
* The `socks` mode of Windows no longer automatically sets the global proxy by default, and the `--setupGlobalProxy` parameter to enable this function is added
* Added `ephemeral` mode for `exchange` command (for k8s 1.23+, thanks to @[xyz-li](https://github.com/xyz-li))
* Fix an issue which cause `exchange` command often stuck (issues #184, thanks to @[xyz-li](https://github.com/xyz-li))
* Provide more elegant error message when the target port of port-forward is occupied (thanks @[xyz-li](https://github.com/xyz-li))
* Automatically control the scope of generated routes according to user permissions, remove the `--global` parameter of the Connect command
* Optimize the `--cidr` parameter of the Connect command to support specifying multiple IP segments
* Parameter `--label` renamed to `--withLabel`
* Added `--withAnnotation` parameter to add extra annotation to shadow pod
* `connect` command adds `--disablePodIp` parameter to support disabling pod IP routing
* shadow pod adds `kt-user` annotation to record local username
* remove `check` command

### 0.1.2

> Release time: 2021-08-29

* Automatically resolve local DNS configuration, remove `--localDomain` parameter of `connect` command
* Automatically detect and install `sshuttle` when using vpn mode, simplifying the preparation for initial use
* Solve the problem of "lost connection to pod" when `exchange` and `mesh` connections are idle and timeout
* Fixed an issue of connection failed when the `connect` command enabled debug mode
* Optimized log output of Windows environment to adapt to non-administrator user scenarios
* Added `--imagePullSecret` parameter to support specifying the Secret used for pulling proxy Pod images (thanks @[pvtyuan](https://github.com/pvtyuan))

### 0.1.1

> Release time: 2021-08-19

* The distribution package is changed from `tar.gz` format to `zip` format, which is convenient for Windows users
* Added `--serviceAccount` parameter to support specifying the ServiceAccount used by the proxy Pod
* Added `--useKubectl` parameter to support using the local `kubectl` tool to connect to the cluster
* Enhanced `clean` command to support cleaning residual ConfigMap and registry data
* Fix issue of Kubernetes address has a context path and cannot be connected
* Fix issue of the owner of the `.ktctl` directory becomes root when executing connect using sudo

### 0.1.0

> Release time: 2021-08-08

* Enhanced `connect` command support under Windows
* Removed dependency on local `kubectl` client tool
* Added `tun` connection mode for Linux (thanks @[xyz-li](https://github.com/xyz-li))
* Use `provide` command instead of `run` command
* Added `clean` command to clean up the remaining shadow pods in the cluster
* Support service domain name resolution of `service.namespace.svc` format
* Improve error message of runtime errors such as missing `sshuttle` dependencies
* `--dump2hosts` parameter of `connect` command supports full service domain name

### 0.0.13-rc13

> Release time: 2021-07-02

* Add `exchange`/`mesh`/`run` plugin for `kubectl` tool
* `exchange` and `mesh` commands support multi-port mapping
* Eliminate local SSH command-line tool dependencies
* Replace fixed delay with port check to improve the execution efficiency of `connect` command
* Support access pod domain name of StatefulSet from local
* Compatible with OpenShift 4.x

### 0.0.12

> Release time: 2020-04-13

* Add `connect` plugin for `kubectl` tool
* Support dump service domain in any namespace to route to the local Hosts file
* Support reuse shadow pod
* Dynamically generate SSH key
* Add`run` command to directly expose the service of the local specified port to the cluster
* Optimized the check waiting for shadow pod to be ready

### 0.0.11

> Release time: 2020-02-27

* Fixed command not exit issue.
* Add `check` command to help check local denpendencies
* Add `dashboard` command to to install and open dashboard in local
* Add support to access service with <servicename>.<namespace>

### 0.0.10

> Release time: 2020-02-02

* Options adaptor windows system
* Add `--dump2hosts` options to support socks5 use


### 0.0.9

> Release time: 2020-01-16

* Support Service Name as dns address
* Make sure shadow is clean up after command exit

### 0.0.8

> Release time: 2019-12-13

* Add windows native support
* Add idea support

### 0.0.7

> Release time: 2019-12-05

* Add oidc plugin to support TKE
* Add socks5 method to support WSL
* Fixed issue when node.Spec.PodCIDR dynamic cal CIDR

### 0.0.6

> Release time: 2019-10-01

* Fixed clusterIP cidr missing
* Rename docker images address

### 0.0.5

> Release time: 2019-10-09

* Add dashboard and api server source code

### 0.0.4

> Release time: 2019-06-26

* Support KtConnect Dashboard

### 0.0.3

> Release time: 2019-06-19

* Add `mesh` command to support istio network rule

### 0.0.2

> Release time: 2019-06-19

* Fixed issue if istio inject is enable in namespace, and the request can't redirect to local
* Support exchange run standalone

### 0.0.1 

> Release time: 2019-06-18

* Split command to `connect` and `exchange`
* Support exchange multiple services
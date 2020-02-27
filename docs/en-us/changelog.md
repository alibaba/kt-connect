## Changelog

### 0.0.11

> Release At 2020-02-27

* fixed command not exit issue.
* add `check` command to help check local denpendencies
* add `dashboard` command to to install and open dashboard in local
* add support to access service with <servicename>.<namespace>

### 0.0.10

> Release At 2020-02-02

* Options adaptor windows system
* Add `--dump2hosts` options to support socks5 use


### 0.0.9

> Release At 2020-01-16

* Support Service Name as dns address
* Make sure shadow is clean up after command exit

### 0.0.8

> Release At 2019-12-13

* Add windows native support
* Add idea support

### 0.0.7

> Release At 2019-12-05

* Add oidc plugin to support TKE
* Add socks5 method to support WSL
* Fixed issue when node.Spec.PodCIRD dynamic cal CIRD

### 0.0.6

> Release At 2019-10-01

* Fixed clusterIP cidr missing.
* Rename docker images address.

### 0.0.5

> Release At 2019-10-09

* Add dashboard and api server source code

### 0.0.4

> Release At 2019-06-26

* Support KT Connect Dashboard

### 0.0.3

> Release At 2019-06-19

* Add `mesh` command to support istio network rule

### 0.0.2

> Release At 2019-06-19

* Fixed issue if istio inject is enable in namespace, and the request can't redirect to local
* Support exchange run standalone.

### 0.0.1 

> Release At 2019-06-18

* split command to `connect` and `exchange`.
* support mutil exchange.
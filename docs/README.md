# Introduction

![](_media/demo-1.gif)

## Features

* Direct access remote Kubernetes cluster.

Developer connect to remote Kubernetes internal network, local development and testing.

* Exchange in cluster request to local.

Developer can exchange the workload to redirect the request to local app.

* Support Service Mesh

You can create a mesh version in local host and redirect to your local

* Light VPN based on SSH

KT use sshuttle as the vpn tool to access remote Kubernetes cluster network.

* As a kubectl plugin

Run kt as a kubectl plugin, all you need is kubectl.

## Release Note

### 0.0.6

* Add windows support doc.
* Fixed clusterIP cidr missing.
* Rename docker images address.

### 0.0.5

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
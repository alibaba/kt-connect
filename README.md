KT-Connect
===========

![Go](https://github.com/alibaba/kt-connect/workflows/Go/badge.svg)
[![Build Status](https://travis-ci.org/alibaba/kt-connect.svg?branch=master)](https://travis-ci.org/alibaba/kt-connect)
[![Go Report Card](https://goreportcard.com/badge/github.com/alibaba/kt-connect)](https://goreportcard.com/report/github.com/alibaba/kt-connect)
[![Test Coverage](https://api.codeclimate.com/v1/badges/eb13b3946784bd7c67cc/test_coverage)](https://codeclimate.com/github/alibaba/kt-connect/test_coverage)
[![Maintainability](https://api.codeclimate.com/v1/badges/eb13b3946784bd7c67cc/maintainability)](https://codeclimate.com/github/alibaba/kt-connect/maintainability)
[![Release](https://img.shields.io/github/release/alibaba/kt-connect.svg?style=flat-square)](https://img.shields.io/github/release/alibaba/kt-connect.svg?style=flat-square)
![License](https://img.shields.io/github/license/alibaba/kt-connect.svg)

KT-Connect (short for "Kubernetes Toolkit Connect") is a utility tool to
manage and integrate with your Kubernetes dev environment more efficiently.

![Arch](./docs/media/arch.png)

## Features

* `Connect`: Directly Access a remote Kubernetes cluster. KtConnect use ssh-vpn, socks-proxy or tun-device to access remote Kubernetes cluster networks.
* `Exchange`: Developer can exchange the workload to redirect the requests to a local app.
* `Mesh`: You can create a mesh version service in local host, and redirect specified workload requests to your local.
* `Provide`: Expose a local running app to Kubernetes cluster as a common service, all requests to that service are redirect to local app.
* `Dashboard`: A dashboard view can help you know how the environment has been used.

## QuickStart

You can download and install the `ktctl` from [Downloads And Install](docs/en-us/downloads.md)

Read the [Quick Start Guide](docs/en-us/quickstart.md) for more about this tool.

## Ask For Help

Please contact us with DingTalk（钉钉）:

<img src="https://raw.githubusercontent.com/alibaba/kt-connect/master/docs/media/dingtalk.png" width="50%"></img>


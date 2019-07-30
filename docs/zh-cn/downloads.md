# 下载和安装

## 二进制包

Mac:

* [Darwin amd64](https://github.com/alibaba/kt-connect/releases/download/v0.0.5/ktctl_darwin_amd64)
* [Darwin 386](https://github.com/alibaba/kt-connect/releases/download/v0.0.5/ktctl_darwin_386)

Linux:

* [Linux Amd64](https://github.com/alibaba/kt-connect/releases/download/v0.0.5/ktctl_linux_amd64)
* [Linux 386](https://github.com/alibaba/kt-connect/releases/download/v0.0.5/ktctl_linux_386)

## Mac用户

安装sshuttle

```
brew install sshuttle
```

下载并安装KT

```
$ curl -OL https://rdc-incubators.oss-cn-beijing.aliyuncs.com/stable/ktctl_darwin_amd64.tar.gz
$ tar -xzvf ktctl_darwin_amd64.tar.gz
$ mv ktctl_darwin_amd64 /usr/local/bin/ktctl
$ ktctl -h
```

## Linux User

安装sshuttle

```
pip install sshuttle
```

下载并安装KT

```
$ curl -OL https://rdc-incubators.oss-cn-beijing.aliyuncs.com/stable/ktctl_linux_amd64.tar.gz
$ tar -xzvf ktctl_linux_amd64.tar.gz
$ mv ktctl_linux_amd64 /usr/local/bin/ktctl
$ ktctl -h
```
Windows Support
==============

> Limit： In Windows environment, you can only use socks5 method to access PodIP and ClusterIP with out DNS.


## Window Native

> Precondition: Please make suere kubectl is already install and can connection to Kubernetes Cluster.

User can download the windows exe biranary from [Daily Build](https://alibaba.github.io/kt-connect/#/nightly). Download and install to PATH.

Exec Command：

```
$ ktctl -d connect --method socks5
```

After this ktctl will create a socks5 proxy in local. and follow the log output set http_proxy environment varibale.

Output:

```
4:31PM INF ==============================================================
4:31PM INF Start SOCKS5 Proxy: export http_proxy=socks5://127.0.0.1:2223
4:31PM INF ==============================================================
```

Set environment:

```
set http_proxy=socks5://127.0.0.1:2223
```

For IDEA User，Please see [How to use in IDEA](https://alibaba.github.io/kt-connect/#/guide/how-to-use-in-idea).

## Windows Subsystem for Linux (WSL)

In order to use kt-connect in Windows you can use WSL in windows 10（released April 2017）。

Flow the guide https://docs.microsoft.com/en-us/windows/wsl/install-win10 and start use kt-connect like in other linux system.

Usage:

```
$ sudo ktctl -d connect --method socks5
4:31PM INF Daemon Start At 56071
4:31PM INF Client address 30.5.125.75
4:31PM INF Deploying shadow deployment kt-connect-daemon-cusdp in namespace default

4:31PM DBG Shadow Pod status is Pending
4:31PM DBG Shadow Pod status is Running
4:31PM INF Shadow is ready.
4:31PM DBG Success deploy proxy deployment kt-connect-daemon-cusdp in namespace default

4:31PM DBG Child, os.Args = [ktctl -d connect --method socks5]
4:31PM DBG Child, cmd.Args = [kubectl --kubeconfig=/Users/yunlong/.kube/config -n default port-forward kt-connect-daemon-cusdp-6d47d4f594-p9m6j 2222:22]
Forwarding from 127.0.0.1:2222 -> 22
Forwarding from [::1]:2222 -> 22
4:31PM DBG port-forward start at pid: 56114
4:31PM INF ==============================================================
4:31PM INF Start SOCKS5 Proxy: export http_proxy=socks5://127.0.0.1:2223
4:31PM INF ==============================================================
4:31PM DBG Child, os.Args = [ktctl -d connect --method socks5]
4:31PM DBG Child, cmd.Args = [ssh -oStrictHostKeyChecking=no -oUserKnownHostsFile=/dev/null -i /Users/yunlong/.kt_id_rsa -D 2223 root@127.0.0.1 -p2222 sh loop.sh]
Handling connection for 2222
Warning: Permanently added '[127.0.0.1]:2222' (ECDSA) to the list of known hosts.
4:31PM DBG vpn(ssh) start at pid: 56190
4:31PM DBG KT proxy start successful
```

KT Connect will create a socks5 proxy in local.

```
# set http_proxy
$ export http_proxy=socks5://127.0.0.1:2223
$ curl http://<POD_IP>:<PORT>
$ curl http://<CLUSTER_IP>:<PORT>
```
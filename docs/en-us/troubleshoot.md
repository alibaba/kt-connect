## Connect Troubleshoot

This chapter wiil guide the user how do set up a connection manually.

### Deploy test application and shadow in cluster

```
$ kubectl run nginx --image=nginx
$ kubectl run troubleshoot --image=registry.cn-hangzhou.aliyuncs.com/rdc-incubator/kt-connect-shadow:stable
```

Make sure the troubleshot pod is ready.

```
$ kubectl get pods -o wide | grep troubleshoot          
NAME                                                              READY   STATUS             RESTARTS   AGE    IP             NODE                                NOMINATED NODE
troubleshoot-6f9958b7f7-g8vgs                                     1/1     Running            0          3m     172.16.1.114
```

### Forward troubleshoot pod port 22 to localhost

```
$ kubectl port-forward pod/troubleshoot-6f9958b7f7-g8vgs 2222:22
Forwarding from 127.0.0.1:2222 -> 22
Forwarding from [::1]:2222 -> 22
```

### Enable ssh with local SSH private key

> the default account of troubleshoot container is `root:root`

```
$ ssh-copy-id root@127.0.0.1 -p 2222
```

Make sure you can login the troubleshoot container from local

```
ssh root@127.0.0.1 -p 2222
```

### Create connection

<!-- tabs:start -->

#### ** Socks5 mode **

Setup socks5 connection：

```
$ ssh -oStrictHostKeyChecking=no -oUserKnownHostsFile=/dev/null -i ~/.ssh/id_rsa -D 5000 root@127.0.0.1 -p 2222
```

Open a new terminal window, and set http_proxy environment variable:

```
$ export http_proxy=socks5://127.0.0.1:5000
$ curl http://<nginx-pod-ip>:80
```

#### ** sshuttle mode **

Setup vpn connection

```
sudo sshuttle -v -e "ssh -oStrictHostKeyChecking=no -oUserKnownHostsFile=/dev/null -i /Users/user/.ssh/id_rsa" -r root@127.0.0.1:2222 -x 127.0.0.1 172.16.0.0/24
```

* -i the absolute path of local ssh private key.
* -x The network segment that needs to be proxy

Open a new termial windows and check connection：

```
$ curl http://<nginx-pod-ip>:80
```

<!-- tabs:end -->



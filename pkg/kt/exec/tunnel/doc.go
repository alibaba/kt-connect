package tunnel

// Copy from `man ssh`
/*
SSH-BASED VIRTUAL PRIVATE NETWORKS
     ssh contains support for Virtual Private Network (VPN) tunnelling using the
     tun(4) network pseudo-device, allowing two networks to be joined securely.
     The sshd_config(5) configuration option PermitTunnel controls whether the
     server supports this, and at what level (layer 2 or 3 traffic).

     The following example would connect client network 10.0.50.0/24 with remote
     network 10.0.99.0/24 using a point-to-point connection from 10.1.1.1 to
     10.1.1.2, provided that the SSH server running on the gateway to the remote
     network, at 192.168.1.15, allows it.

     On the client:

           # ssh -f -w 0:1 192.168.1.15 true
           # ifconfig tun0 10.1.1.1 10.1.1.2 netmask 255.255.255.252
           # route add 10.0.99.0/24 10.1.1.2

     On the server:

           # ifconfig tun1 10.1.1.2 10.1.1.1 netmask 255.255.255.252
           # route add 10.0.50.0/24 10.1.1.1

     Client access may be more finely tuned via the /root/.ssh/authorized_keys
     file (see below) and the PermitRootLogin server option.  The following entry
     would permit connections on tun(4) device 1 from user ``jane'' and on tun
     device 2 from user ``john'', if PermitRootLogin is set to
     ``forced-commands-only'':

       tunnel="1",command="sh /etc/netstart tun1" ssh-rsa ... jane
       tunnel="2",command="sh /etc/netstart tun2" ssh-rsa ... john

     Since an SSH-based setup entails a fair amount of overhead, it may be more
     suited to temporary setups, such as for wireless VPNs.  More permanent VPNs
     are better provided by tools such as ipsecctl(8) and isakmpd(8).

*/

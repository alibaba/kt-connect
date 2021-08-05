#!/bin/bash
mkdir -p /root/.ssh
if [ -r /root/authorized/authorized_keys ]; then
    cp /root/authorized/authorized_keys /root/.ssh
fi

/usr/sbin/sshd -D &

del_device_handler() {
  echo "delete tun device tun0"
  ip l d tun0
}

if [ -n "${CLIENT_TUN_IP}" -a -n "${SERVER_TUN_IP}" ]; then
  trap 'del_device_handler' SIGTERM
  echo "tun mod, client tun ip: ${CLIENT_TUN_IP}, local tun ip: ${SERVER_TUN_IP} "
  echo "create tun device tun1"
  ip tuntap add dev tun1 mod tun
  echo "setup device ip"
  ip address add "${SERVER_TUN_IP}"/"${TUN_MASK_LEN}" dev tun1
  echo "turn device up"
  ip link set dev tun1 up

  echo "set up iptables"
  iptables -t nat -A POSTROUTING -s "${CLIENT_TUN_IP}" -o eth0 -j MASQUERADE
fi

if [[ "${1}" = "--debug" ]]; then
   /usr/sbin/dlv --listen=:2345 --headless=true --api-version=2 --accept-multiclient exec /usr/sbin/shadow-linux-amd64
else
   /usr/sbin/shadow-linux-amd64
fi

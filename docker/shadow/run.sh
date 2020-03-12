#!/bin/bash
mkdir -p /root/.ssh
cp /root/authorized/authorized_keys /root/.ssh

/usr/sbin/sshd -D &

if [[ "${1}" = "--debug" ]]; then
    /usr/sbin/dlv --listen=:2345 --headless=true --api-version=2 --accept-multiclient exec /usr/sbin/shadow-linux-amd64
else
    /usr/sbin/shadow-linux-amd64
fi

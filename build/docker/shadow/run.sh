#!/bin/bash

echo "shadow starting ..."

# fetch authorized_keys from volume mounted via config map
mkdir -p /root/.ssh
cp /root/authorized/authorized_keys /root/.ssh/authorized_keys

if [ -n "${privateKey}" ]; then
  # for ephemeral container
  # private key and authorized_keys must be base64 encoded in environment
  echo "${privateKey}" | base64 -d > /root/.ssh/id_rsa
  echo "private key created created"
fi

/usr/sbin/sshd -D &
echo "sshd triggered"

sleep 1
ps -ef | grep -v 'ps -ef' | cat

if [[ "${1}" = "--debug" ]]; then
  echo "run in debug mode"
  /usr/sbin/dlv --listen=:2345 --headless=true --api-version=2 --accept-multiclient exec /usr/sbin/shadow
else
  echo "run in standard mode"
  /usr/sbin/shadow
fi

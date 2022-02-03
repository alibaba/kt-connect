#!/bin/bash

echo "Initializing ..."
env | grep 'KT_'

# fetch authorized_keys from volume mounted via config map
mkdir -p /root/.ssh
cp /root/authorized/authorized_keys /root/.ssh/authorized_keys

if [ -n "${privateKey}" ]; then
  # for ephemeral container
  # private key and authorized_keys must be base64 encoded in environment
  echo "${privateKey}" | base64 -d > /root/.ssh/id_rsa
  echo "Private key created created"
fi

if [ "${KT_DNS_PROTOCOL}" = "" ]; then
  echo "Skip shadow process"
elif [ "${1}" = "--debug" ]; then
  echo "Run shadow in debug mode"
  /usr/sbin/dlv --listen=:2345 --headless=true --api-version=2 --accept-multiclient exec /usr/sbin/shadow &
else
  echo "Run shadow in standard mode"
  /usr/sbin/shadow &
fi

/usr/sbin/sshd -D
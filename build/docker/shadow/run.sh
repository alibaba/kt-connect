#!/bin/bash
mkdir -p /root/.ssh

# mounted from config map
cp /root/authorized/authorized_keys /root/.ssh

if [ -n "${privateKey}" ]; then
  # for ephemeral container
  # private key and authorized_keys must be base64 encoded in environment
  echo "${privateKey}" |base64 -d > /root/.ssh/id_rsa
fi

/usr/sbin/sshd -D &

if [[ "${1}" = "--debug" ]]; then
   /usr/sbin/dlv --listen=:2345 --headless=true --api-version=2 --accept-multiclient exec /usr/sbin/shadow
else
   /usr/sbin/shadow
fi

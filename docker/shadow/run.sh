#!/bin/bash
/usr/sbin/sshd -D &

if [[ "${1}" = "--debug" ]]; then
    /usr/sbin/dlv --listen=:40000 --headless=true --api-version=2 exec /usr/sbin/kt-shadow
else
    /usr/sbin/kt-shadow
fi

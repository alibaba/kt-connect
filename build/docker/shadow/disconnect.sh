#!/bin/bash

port=${1}
if [ "${port}" = "" ]; then
  echo "must specify a port to disconnect"
  exit 1
fi
echo "disconnect port ${port}"

pid=$(lsof -i :"${port}" | grep 'IPv4' | head -1 | awk '{print $2}')
if [ "${pid}" = "" ]; then
  echo "no process using port ${port}"
  exit 0
fi
echo "process ${pid} is using port ${port}"

if ! kill -15 ${pid} 2>&1; then
  echo "fail to disconnected port ${port}"
else
  echo "port ${port} disconnected"
fi
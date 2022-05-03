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

kill -15 ${pid} 2>&1
if [ ${?} -eq 0 ]; then
  echo "port ${port} disconnected"
else
  echo "fail to disconnected port ${port}"
fi
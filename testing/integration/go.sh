#!/usr/bin/env bash

NS="kt-integration-test"
IMAGE="registry.cn-hangzhou.aliyuncs.com/rdc-incubator/kt-connect-shadow:latest"
MODE="vpn"

# Log
function log() {
    printf ">> ${*}\n"
}

# Clean everything up
function cleanup() {
  log "cleaning up ..."
  for i in `jobs -l | grep '^\[[0-9]\]' | cut -c 1-30 | grep -o ' [0-9]\+ ' | awk {'print $1'}`; do sudo kill $i; done
  kubectl -n ${NS} delete deployment tomcat
  kubectl -n ${NS} delete service tomcat
  kubectl delete namespace ${NS}
}

# Exit with error
function fail() {
  log "'\e[31m${*} !!!\e[0m"
  cleanup
  exit -1
}

# Wait pod ready
function wait_for_pod() {
  for i in `seq 10`; do
    log "checking pod ${1}, ${i} times"
    exist=`kubectl -n ${NS} get pod | grep "^${1}-" | grep "${2}/${2}"`
    if [ "$exist" != "" ]; then
      sleep 2
      return
    fi
    sleep 3
  done
  fail "failed to start pod ${1}"
}

# Require for root access
sudo true
if [ ${?} -ne 0 ]; then fail "failed to require root access"; fi

# Check environment is clean
existPid=`ps aux | grep "ktctl" | grep -v "grep" | awk '{print $2}' | sort -n | head -1`
if [ "${existPid}" != "" ]; then fail "ktctl already running before test start (pid: ${existPid})"; fi

ktctl clean >/dev/null 2>&1

# Prepare test resources
kubectl create namespace ${NS}

kubectl -n ${NS} create deployment tomcat --image=tomcat:9 --port=8080
kubectl -n ${NS} expose deployment tomcat --port=8080 --target-port=8080
wait_for_pod tomcat 1
kubectl -n ${NS} exec deployment/tomcat -c tomcat -- /bin/bash -c 'mkdir webapps/ROOT; echo "kt-connect demo v1" > webapps/ROOT/index.html'

podIp=`kubectl -n ${NS} get pod --selector app=tomcat -o jsonpath='{.items[0].status.podIP}'`
log "tomcat pod-ip: ${podIp}"
if [ ${podIp} = "" ]; then fail "failed to setup test deployment"; fi
clusterIP=`kubectl -n ${NS} get service tomcat -o jsonpath='{.spec.clusterIP}'`
log "tomcat cluster-ip: ${clusterIP}"
if [ ${clusterIP} = "" ]; then fail "failed to setup test service"; fi

# Test connect
sudo ktctl -n ${NS} -i ${IMAGE} -f connect --method ${MODE} >/tmp/kt-ci-connect.log 2>&1 &
sleep 1

declare -i count=`jobs | grep 'connect' | wc -l`
if [ ${count} -ne 1 ]; then fail "failed to setup ktctl connect"; fi

pidFile=`ls -t ~/.ktctl/connect-*.pid | less -1`
connectPid=`cat ${pidFile}`
log "ktctl connect pid: ${connectPid}"

res=`curl -s "http://${podIp}:8080"`
if [ "${res}" != "kt-connect demo v1" ]; then fail "failed to access via pod-ip, got: ${res}"; fi
res=`curl -s "http://${clusterIP}:8080"`
if [ "${res}" != "kt-connect demo v1" ]; then fail "failed to access via cluster-ip, got: ${res}"; fi
res=`curl -s "http://tomcat.${NS}.svc.cluster.local:8080"`
if [ "${res}" != "kt-connect demo v1" ]; then fail "failed to access via service-domain, got: ${res}"; fi

# Prepare local service
while true; do
  printf "HTTP/1.1 200 OK\nContent-Length: 19\nContent-Type: text/plain\n\nkt-connect local v2" | nc -l 8080
  sleep 1
done >/tmp/kt-ci-nc.log 2>&1 &
sleep 1

res=`curl -s "http://127.0.0.1:8080"`
if [ "${res}" != "kt-connect local v2" ]; then fail "failed to setup local service, got: ${res}"; fi

ncPid=`jobs -l | grep 'while true' | grep -o ' [0-9]\+ ' | awk {'print $1'}`
log "local server pid: ${ncPid}"

# Test exchange
ktctl -n ${NS} -i ${IMAGE} -f exchange tomcat --expose 8080 >/tmp/kt-ci-exchange.log 2>&1 &
sleep 1

declare -i count=`jobs | grep 'exchange' | wc -l`
if [ ${count} -ne 1 ]; then fail "failed to setup ktctl exchange"; fi

pidFile=`ls -t ~/.ktctl/exchange-*.pid | less -1`
exchangePid=`cat ${pidFile}`
log "ktctl exchange pid: ${exchangePid}"
wait_for_pod tomcat-kt 1

res=`curl -s "http://tomcat.${NS}.svc.cluster.local:8080"`
if [ "${res}" != "kt-connect local v2" ]; then fail "failed to exchange test service, got: ${res}"; fi

log "\e[32mall tests done !!!\e[0m"

# Clean up
cleanup

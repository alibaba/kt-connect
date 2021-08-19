#!/usr/bin/env bash

NS="kt-integration-test"
IMAGE="registry.cn-hangzhou.aliyuncs.com/rdc-incubator/kt-connect-shadow:latest"
MODE="vpn"
DOCKER_HOST="ubuntu@192.168.64.2"

# Log
function log() {
    printf "\e[34m>>\e[0m ${*}\n"
}

# Clean everything up
function cleanup() {
  log "cleaning up ..."
  for i in `jobs -l | grep '^\[[0-9]\]' | cut -c 1-30 | grep -o ' [0-9]\+ ' | awk {'print $1'}`; do
    log "killing process ${i}"
    sudo kill -15 ${i};
  done
  if [ "${DOCKER_HOST}" != "" ]; then
    PID=`ps aux | grep 'ssh -CfNgL 8080:localhost:8080' | grep -v 'grep' | awk '{print $2}'`
    if [ "${PID}" != "" ]; then
      kill -15 ${PID}
    fi
  fi
  docker rm -f tomcat
  kubectl -n ${NS} delete deployment tomcat
  kubectl -n ${NS} delete service tomcat
  kubectl delete namespace ${NS}
}

# Exit with error
function fail() {
  log "\e[31m${*} !!!\e[0m"
  log "check logs for detail: \e[33m`ls -t /tmp/kt-it-*.log`\e[0m"
  cleanup
  exit -1
}

# Test passed
function success() {
  log "\e[32m${*} !!!\e[0m"
}

# Wait pod ready
function wait_for_pod() {
  for i in `seq 10`; do
    log "checking pod ${1}, ${i} times"
    exist=`kubectl -n ${NS} get pod | grep "^${1}-" | grep "${2}/${2}"`
    if [ "$exist" != "" ]; then
      return
    fi
    sleep 3
  done
  fail "failed to start pod ${1}"
}

# Check if background job running
function check_job() {
    declare -i count=`jobs | grep "${1}" | wc -l`
    if [ ${count} -ne 1 ]; then fail "failed to setup ${1} job"; fi
}

# Check if ktctl pid file exists
function check_pid_file() {
    pidFile=`ls -t ~/.ktctl/${1}-*.pid | less -1`
    pid=`cat ${pidFile}`
    log "ktctl ${1} pid: ${pid}"
}

# Verify access specified url with result
function verify() {
  target=${1}
  url=${2}
  shift 2
  log "accessing ${url}"
  for c in `seq 5`; do
    res=`curl --connect-timeout 2 -s ${url}`
    if [ "$res" = "${*}" ]; then
      return
    fi
    log "retry times: ${c}"
    sleep 2
  done
  fail "failed to access ${target}, got: ${res}"
}

# Require for root access
sudo true
if [ ${?} -ne 0 ]; then fail "failed to require root access"; fi

# Check environment is clean
existPid=`ps aux | grep "ktctl" | grep -v "grep" | awk '{print $2}' | sort -n | head -1`
if [ "${existPid}" != "" ]; then fail "ktctl already running before test start (pid: ${existPid})"; fi

rm -f /tmp/kt-it-*.log
ktctl -n ${NS} clean >/dev/null 2>&1

# Prepare test resources
kubectl create namespace ${NS}

kubectl -n ${NS} create deployment tomcat --image=tomcat:9 --port=8080
kubectl -n ${NS} expose deployment tomcat --port=8080 --target-port=8080
wait_for_pod tomcat 1
kubectl -n ${NS} exec deployment/tomcat -c tomcat -- /bin/bash -c 'mkdir -p webapps/ROOT; echo "kt-connect demo v1" > webapps/ROOT/index.html'

podIp=`kubectl -n ${NS} get pod --selector app=tomcat -o jsonpath='{.items[0].status.podIP}'`
log "tomcat pod-ip: ${podIp}"
if [ ${podIp} = "" ]; then fail "failed to setup test deployment"; fi
clusterIP=`kubectl -n ${NS} get service tomcat -o jsonpath='{.spec.clusterIP}'`
log "tomcat cluster-ip: ${clusterIP}"
if [ ${clusterIP} = "" ]; then fail "failed to setup test service"; fi

# Test connect
sudo ktctl -n ${NS} -i ${IMAGE} -f connect --method ${MODE} >/tmp/kt-it-connect.log 2>&1 &
wait_for_pod kt-connect 1
check_job connect
check_pid_file connect

verify "pod-ip" "http://${podIp}:8080" "kt-connect demo v1"
verify "cluster-ip" "http://${clusterIP}:8080" "kt-connect demo v1"
verify "service-domain-full-qualified" "http://tomcat.${NS}.svc.cluster.local:8080" "kt-connect demo v1"
verify "service-domain-with-namespace" "http://tomcat.${NS}:8080" "kt-connect demo v1"
verify "service-domain" "http://tomcat:8080" "kt-connect demo v1"
success "ktctl connect test passed"

# Prepare local service
docker run -d --name tomcat -p 8080:8080 tomcat:9
sleep 1

exist=`docker ps -a | grep ' tomcat$' | grep -i ' Up '`
if [ "${exist}" = "" ]; then fail "failed to start up local tomcat container"; fi
docker exec tomcat /bin/bash -c 'mkdir webapps/ROOT; echo "kt-connect local v2" > webapps/ROOT/index.html'

if [ "${DOCKER_HOST}" != "" ]; then
  ssh -CfNgL 8080:localhost:8080 ${DOCKER_HOST}
fi

verify "local-service" "http://127.0.0.1:8080" "kt-connect local v2"

# Test exchange
ktctl -n ${NS} -i ${IMAGE} -f exchange tomcat --expose 8080 >/tmp/kt-it-exchange.log 2>&1 &
wait_for_pod tomcat-kt 1
check_job exchange
check_pid_file exchange

verify "service-domain" "http://tomcat.${NS}.svc.cluster.local:8080" "kt-connect local v2"
success "ktctl exchange test passed"

# Test provide
ktctl -n ${NS} -i ${IMAGE} -f provide tomcat-preview --expose 8080 >/tmp/kt-it-provide.log 2>&1 &
wait_for_pod tomcat-preview-kt 1
check_job provide
check_pid_file provide

verify "service-domain" "http://tomcat-preview.${NS}.svc.cluster.local:8080" "kt-connect local v2"
success "ktctl provide test passed"

# Clean up
cleanup
success "all tests done"

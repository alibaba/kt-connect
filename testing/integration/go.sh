#!/usr/bin/env bash

NS="kt-integration-test"
SHADOW_IMAGE="registry.cn-hangzhou.aliyuncs.com/rdc-incubator/kt-connect-shadow:vdev"
ROUTER_IMAGE="registry.cn-hangzhou.aliyuncs.com/rdc-incubator/kt-connect-router:vdev"
DOCKER_HOST="ubuntu@192.168.64.2"
DNS_MODE="localDNS"
CONNECT_MODE="tun2socks"
EXCHANGE_MODE="scale"
MESH_MODE="auto"
declare -i RETRY_TIMES=10
CLEANUP_ONLY="N"
KEEP_PROOF="N"

# Print usage
function usage() {
  echo "go.sh [--keep-proof] [--cleanup-only]"
}

# Log
function log() {
  printf "\e[34m>>\e[0m ${*}\n"
}

# Exit with error
function fail() {
  error "${@}"
  if [ "${KEEP_PROOF}" != "Y" ]; then
    cleanup
  fi
  exit 1
}

# Print error message
function error() {
  log "\e[31m${*} !!!\e[0m"
  log "check logs for detail: \e[33m`ls -cr /tmp/kt-it-*.log`\e[0m"
}

# Test passed
function success() {
  log "\e[32m${*} !!!\e[0m"
}

# Clean everything up
function cleanup() {
  log "cleaning up ..."
  rm -f ${HOME}/.kt/pid/*.pid
  if [ "${DOCKER_HOST}" != "" ]; then
    PID=`ps aux | grep 'CfNgL 8080:localhost:8080' | grep -v 'grep' | awk '{print $2}'`
    if [ "${PID}" != "" ]; then
      log "disconnect from docker host ${DOCKER_HOST}"
      kill -15 ${PID}
    fi
  fi
  docker rm -f tomcat
  kubectl -n ${NS} delete deployment tomcat
  kubectl -n ${NS} delete service tomcat
  check_resources_cleaned
  kubectl delete namespace ${NS}
}

# Wait all resource created by ktctl get cleaned
function check_resources_cleaned() {
  for i in `seq 10`; do
    sleep 6
    log "checking resource clean up, ${i} times"
    resource_count=`kubectl -n ${NS} get pod,configmap,service | grep '^\(pod/\|configmap/\|service/\).*' | grep -v 'kube-root-ca\.crt' | wc -l`
    if [ $resource_count -eq 0 ]; then
      return
    fi
  done
  kubectl -n ${NS} get pod,configmap,service -o wide >/tmp/kt-it-left.log
  error "some resource are left in namespace ${NS}"
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
  pidFile=`ls -t ${HOME}/.kt/pid/${1}-*.pid | head -1`
  pid=`cat ${pidFile}`
  log "ktctl ${1} pid: ${pid}"
}

# Verify access specified url with result
function verify() {
  target="${1}"
  url="${2}"
  if [ "${url:0:4}" != "http" ]; then
    header="${2}"
    url="${3}"
    shift 3
  else
    shift 2
  fi
  log "accessing ${url}"
  for c in `seq ${RETRY_TIMES}`; do
    if [ "${header}" != "" ]; then
      res=`curl -H "${header}" --connect-timeout 2 -s ${url}`
    else
      res=`curl --connect-timeout 2 -s ${url}`
    fi
    if [ "$res" = "${*}" ]; then
      return
    fi
    log "retry times: ${c}"
    sleep 5
  done
  fail "failed to access ${target}, got: ${res}"
}

function prepare_cluster() {
  # Require for root access
  sudo true
  if [ ${?} -ne 0 ]; then fail "failed to require root access"; fi

  # Check environment is clean
  existPid=`ps aux | grep "ktctl" | grep -v "grep" | awk '{print $2}' | sort -n | head -1`
  if [ "${existPid}" != "" ]; then fail "ktctl already running before test start (pid: ${existPid})"; fi

  rm -f /tmp/kt-it-*.log
  ktctl -n ${NS} clean >/tmp/kt-it-clean.log 2>&1
  if [ ${?} -ne 0 ]; then fail "clean up failed (kubernetes cluster unreachable ?)"; fi

  # Prepare test resources
  kubectl create namespace ${NS}

  kubectl -n ${NS} create deployment tomcat --image=tomcat:9 --port=8080
  kubectl -n ${NS} expose deployment tomcat --port=8080 --target-port=8080
  wait_for_pod tomcat 1
  kubectl -n ${NS} exec deployment/tomcat -c tomcat -- /bin/bash \
    -c 'mkdir -p webapps/ROOT; echo "kt-connect demo v1" > webapps/ROOT/index.html'

  podIp=`kubectl -n ${NS} get pod --selector app=tomcat -o jsonpath='{.items[0].status.podIP}'`
  log "tomcat pod-ip: ${podIp}"
  if [ "${podIp}" = "" ]; then fail "failed to setup test deployment"; fi
  clusterIP=`kubectl -n ${NS} get service tomcat -o jsonpath='{.spec.clusterIP}'`
  log "tomcat cluster-ip: ${clusterIP}"
  if [ "${clusterIP}" = "" ]; then fail "failed to setup test service"; fi
}

function test_ktctl_connect() {
  # Test connect
  if [ "${DOCKER_HOST}" == "" ]; then
    sudo ktctl -d -n ${NS} -i ${SHADOW_IMAGE} -f connect --mode ${CONNECT_MODE} --dnsMode ${DNS_MODE} --dnsCacheTtl 10 >/tmp/kt-it-connect.log 2>&1 &
  else
    sudo ktctl -d -n ${NS} -i ${SHADOW_IMAGE} -f connect --mode ${CONNECT_MODE} --dnsMode ${DNS_MODE} --dnsCacheTtl 10 --excludeIps ${DOCKER_HOST#*@} >/tmp/kt-it-connect.log 2>&1 &
  fi
  wait_for_pod kt-connect 1
  check_job connect
  check_pid_file connect

  verify "pod-ip" "http://${podIp}:8080" "kt-connect demo v1"
  verify "cluster-ip" "http://${clusterIP}:8080" "kt-connect demo v1"
  verify "service-domain-full-qualified" "http://tomcat.${NS}.svc.cluster.local:8080" "kt-connect demo v1"
  verify "service-domain-with-namespace" "http://tomcat.${NS}:8080" "kt-connect demo v1"
  verify "service-domain" "http://tomcat:8080" "kt-connect demo v1"
  success "ktctl connect test passed"
}

function prepare_local() {
  # Prepare local service
  docker run -d --name tomcat -p 8080:8080 tomcat:9
  if [ $? -eq 0 ]; then
    log "local tomcat container started"
  else
    fail "failed to start local tomcat container"
  fi
  sleep 3

  exist=`docker ps -a | grep ' tomcat$' | grep -i ' Up '`
  if [ "${exist}" = "" ]; then fail "failed to start up local tomcat container"; fi
  docker exec tomcat /bin/bash -c 'mkdir -p webapps/ROOT; echo "kt-connect local v2" > webapps/ROOT/index.html'
  if [ $? -ne 0 ]; then fail "failed to update tomcat index page content"; fi

  if [ "${DOCKER_HOST}" != "" ]; then
    ssh -o "UserKnownHostsFile=/dev/null" -o "StrictHostKeyChecking=no" -CfNgL 8080:localhost:8080 ${DOCKER_HOST}
  fi

  verify "local-service" "http://127.0.0.1:8080" "kt-connect local v2"
}

function test_ktctl_exchange() {
  # Test exchange
  ktctl -d -n ${NS} -i ${SHADOW_IMAGE} -f exchange deployment/tomcat --mode ${EXCHANGE_MODE} --expose 8080 >/tmp/kt-it-exchange.log 2>&1 &
  wait_for_pod tomcat-kt-exchange 1
  check_job exchange
  check_pid_file exchange

  verify "exchanged-service" "http://tomcat.${NS}.svc.cluster.local:8080" "kt-connect local v2"
  success "ktctl exchange test passed"
}

function test_ktctl_mesh() {
  # Test mesh
  ktctl -d -n ${NS} -i ${SHADOW_IMAGE} -f mesh tomcat --mode ${MESH_MODE} --expose 8080 --versionMark ci \
    --routerImage ${ROUTER_IMAGE} >/tmp/kt-it-mesh.log 2>&1 &
  wait_for_pod tomcat-kt-mesh 1
  check_job mesh
  check_pid_file mesh

  verify "without-header" "http://tomcat.${NS}.svc.cluster.local:8080" "kt-connect demo v1"
  verify "with-header" "VERSION:ci" "http://tomcat.${NS}.svc.cluster.local:8080" "kt-connect local v2"
  success "ktctl mesh test passed"
}

function test_ktctl_preview() {
  # Test preview
  ktctl -d -n ${NS} -i ${SHADOW_IMAGE} -f preview tomcat-preview --expose 8080 >/tmp/kt-it-preview.log 2>&1 &
  wait_for_pod tomcat-preview-kt 1
  check_job preview
  check_pid_file preview

  sleep 3
  verify "service-domain" "http://tomcat-preview.${NS}.svc.cluster.local:8080" "kt-connect local v2"
  success "ktctl preview test passed"
}

if [ "${1}" = "--help" ]; then
  usage
  exit 0
elif [ "${1}" = "--cleanup-only" ]; then
  CLEANUP_ONLY="Y"
elif [ "${1}" = "--keep-proof" ]; then
  KEEP_PROOF="Y"
fi

if [ "${CLEANUP_ONLY}" = "Y" ]; then
  cleanup
  success "cleanup done"
else
  prepare_cluster
  test_ktctl_connect
  prepare_local
  test_ktctl_mesh
  test_ktctl_exchange
  test_ktctl_preview
  cleanup
  success "all tests done"
fi

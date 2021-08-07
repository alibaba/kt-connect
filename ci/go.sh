#!/usr/bin/env bash

NS="kt-ci-test"
IMAGE="registry.cn-hangzhou.aliyuncs.com/rdc-incubator/kt-connect-shadow:latest"

# Exit with error
function fail() {
  echo "${*}"
  exit -1
}

# Wait pod ready
function wait_for_pod() {
  for i in `seq 10`; do
    echo "checking pod ${1}, ${i} times"
    exist=`kubectl -n ${NS} get pod | grep "^${1}-" | grep "${2}/${2}"`
    if [ "$exist" != "" ]; then
      return
    fi
    sleep 3
  done
  fail "failed to start pod ${1}"
}

# Check environment is clean
declare -i count=`ps aux | grep "ktctl" | grep -v "grep" | wc -l`
if [ ${count} -ne 0 ]; then fail "ktctl already running before test start"; fi

# Prepare test resources
kubectl create namespace ${NS}

kubectl -n ${NS} create deployment tomcat --image=tomcat:9 --port=8080
kubectl -n ${NS} expose deployment tomcat --port=8080 --target-port=8080
wait_for_pod tomcat 1
kubectl -n ${NS} exec deployment/tomcat -c tomcat -- /bin/bash -c 'mkdir webapps/ROOT; echo "kt-connect demo v1" > webapps/ROOT/index.html'

podIp=`kubectl -n ${NS} get pod --selector app=tomcat -o jsonpath='{.items[0].status.podIP}'`
echo "tomcat pod-ip: ${podIp}"
if [ ${podIp} = "" ]; then fail "failed to setup test deployment"; fi
clusterIP=`kubectl -n ${NS} get service tomcat -o jsonpath='{.spec.clusterIP}'`
echo "tomcat cluster-ip: ${clusterIP}"
if [ ${clusterIP} = "" ]; then fail "failed to setup test service"; fi

# Test connect
sudo true
sudo ktctl -n ${NS} -i ${IMAGE} -f connect --method vpn >/tmp/kt-ci-connect.log 2>&1 &
sleep 1

declare -i count=`jobs | grep 'connect' | wc -l`
if [ ${count} -ne 1 ]; then fail "failed to setup ktctl connect"; fi

pidFile=`ls -t ~/.ktctl/connect-*.pid | less -1`
connectPid=`cat ${pidFile}`
echo "ktctl connect pid: ${connectPid}"

res=`curl -s "http://${podIp}:8080"`
if [ ${res} != "kt-connect demo v1" ]; then fail "failed to access via pod-ip, got ${res}"; fi
res=`curl -s "http://${clusterIP}:8080"`
if [ ${res} != "kt-connect demo v1" ]; then fail "failed to access via cluster-ip, got ${res}"; fi
res=`curl -s "http://tomcat.${NS}.svc.cluster.local:8080"`
if [ ${res} != "kt-connect demo v1" ]; then fail "failed to access via service-domain, got ${res}"; fi

# Prepare local service
docker run -d --name tomcat -p 8080:8080 tomcat:9
sleep 1
docker exec tomcat /bin/bash -c 'mkdir webapps/ROOT; echo "kt-connect local v2" > webapps/ROOT/index.html'

res=`curl -s "http://127.0.0.1:8080"`
if [ ${res} != "kt-connect local v2" ]; then fail "failed to setup local service, got ${res}"; fi

# Test exchange
ktctl -n ${NS} -i ${IMAGE} -f exchange tomcat --expose 8080 >/tmp/kt-ci-exchange.log 2>&1 &
sleep 1

declare -i count=`jobs | grep 'exchange' | wc -l`
if [ ${count} -ne 1 ]; then fail "failed to setup ktctl exchange"; fi

pidFile=`ls -t ~/.ktctl/exchange-*.pid | less -1`
exchangePid=`cat ${pidFile}`
echo "ktctl exchange pid: ${exchangePid}"

res=`curl -s "http://tomcat.${NS}.svc.cluster.local:8080"`
if [ "${res}" != "kt-connect local v2" ]; then fail "failed to exchange test service, got ${res}"; fi

# Clean up
sudo kill -15 ${exchangePid}
docker rm -f tomcat
sudo kill -15 ${connectPid}
kubectl -n ${NS} delete deployment tomcat
kubectl -n ${NS} delete service tomcat
kubectl delete namespace ${NS}

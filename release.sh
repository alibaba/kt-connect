docker build -t registry.cn-hangzhou.aliyuncs.com/rdc-incubator/kt-connect-shadow -f docker/shadow/Dockerfile .
docker push registry.cn-hangzhou.aliyuncs.com/rdc-incubator/kt-connect-shadow

docker build -t registry.cn-hangzhou.aliyuncs.com/rdc-incubator/kt-connect -f docker/kt-connect/Dockerfile .
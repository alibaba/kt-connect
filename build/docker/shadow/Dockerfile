FROM registry.cn-hangzhou.aliyuncs.com/rdc-incubator/shadow-base:v0.1.0
COPY artifacts/shadow/shadow-linux-amd64 /usr/sbin/shadow-linux-amd64
COPY build/docker/shadow/run.sh /run.sh
RUN chmod 755 /run.sh

ENTRYPOINT ["/run.sh"]

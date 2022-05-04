FROM registry.cn-hangzhou.aliyuncs.com/rdc-incubator/kt-shadow-base:v0.4.0

COPY artifacts/shadow/shadow-linux-amd64 /usr/sbin/shadow
COPY build/docker/shadow/run.sh /run.sh
COPY build/docker/shadow/disconnect.sh /disconnect.sh

RUN chmod 755 /disconnect.sh && \
    chmod 755 /run.sh

ENTRYPOINT ["/run.sh"]

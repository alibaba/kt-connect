FROM registry.cn-hangzhou.aliyuncs.com/rdc-incubator/shadow-base:v0.3.0

COPY artifacts/shadow/shadow-linux-amd64 /usr/sbin/shadow
COPY build/docker/shadow/run.sh /run.sh
COPY build/docker/shadow/setup_iptables.sh /setup_iptables.sh

RUN chmod 755 /setup_iptables.sh && \
    chmod 755 /run.sh

ENTRYPOINT ["/run.sh"]

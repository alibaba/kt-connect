FROM registry.cn-hangzhou.aliyuncs.com/rdc-incubator/shadow-base:v0.1.13
COPY artifacts/shadow/shadow-linux-amd64 /usr/sbin/shadow-linux-amd64
RUN apt-get install -y net-tools iproute2
RUN echo "\nPermitTunnel yes" >> /etc/ssh/sshd_config
ADD build/docker/shadow/run.sh /run.sh
RUN chmod 755 /run.sh

ENTRYPOINT ["/run.sh"]

FROM openresty/openresty:1.21.4.1rc3-bullseye-fat

COPY artifacts/navigator/navigator-linux-amd64 /usr/sbin/navigator
COPY build/docker/navigator/setup_iptables.sh /setup_iptables.sh

RUN chmod 755 /setup_iptables.sh

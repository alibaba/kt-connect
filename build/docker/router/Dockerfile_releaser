FROM nginx:1.21

COPY build/docker/router/nginx.conf /etc/nginx/nginx.conf
COPY router /usr/sbin/router

RUN rm -f /etc/nginx/conf.d/*.conf && \
    chmod +x /usr/sbin/router && \
    touch /var/kt.lock

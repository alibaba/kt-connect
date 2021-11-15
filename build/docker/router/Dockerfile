FROM nginx:1.21.4-alpine

COPY build/docker/router/nginx.conf /etc/nginx/nginx.conf
RUN rm -f /etc/nginx/conf.d/*.conf

COPY artifacts/router/router-linux-amd64 /usr/sbin/router

FROM node:14.8.0 as build-deps
WORKDIR /usr/src/app
COPY web/package.json ./
RUN npm install -g cnpm 
RUN cnpm install
COPY web ./
RUN cnpm run build

FROM nginx:1.15
ENV APP_ENV PROD
COPY --from=build-deps /usr/src/app/build /usr/share/nginx/html
ADD build/docker/dashboard/nginx/default.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
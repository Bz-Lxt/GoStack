FROM nginx:alpine
RUN set -e; \
    sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories || true; \
    apk add --no-cache wget tzdata || apk add --no-cache --repository https://dl-cdn.alpinelinux.org/alpine/latest-stable/main wget tzdata
ENV TZ=Asia/Shanghai
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY dist /usr/share/nginx/html
EXPOSE 80

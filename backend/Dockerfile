FROM alpine:3.21
RUN set -e; \
    sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories || true; \
    apk add --no-cache iproute2 wget ca-certificates tzdata || \
    apk add --no-cache iproute2 wget ca-certificates tzdata; \
    adduser -D -u 1000 gostack
ENV TZ=Asia/Shanghai
COPY gostack /usr/local/bin/gostack
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh /usr/local/bin/gostack
EXPOSE 8080
ENTRYPOINT ["/entrypoint.sh"]

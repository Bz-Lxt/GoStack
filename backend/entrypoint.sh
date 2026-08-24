#!/bin/sh
set -e
export TZ="${TZ:-Asia/Shanghai}"
if [ ! -e /dev/net/tun ]; then
  mkdir -p /dev/net
  mknod /dev/net/tun c 10 200 || true
  chmod 666 /dev/net/tun || true
fi
exec /usr/local/bin/gostack "$@"

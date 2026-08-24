# GoStack 评测说明

本项目是基于Go语言的基于 TUN/TAP 驱动的自研高性能用户态 TCP 协议栈与网络流量拓扑大屏（Mini GoStack），旨在解决基于 TUN/TAP 驱动的自研高性能用户态 TCP 协议栈与网络流量拓扑大屏（Mini GoStack）相关的工程问题，使用了Go、Vue 3，功能有TCP 状态机动态翻转墙、数据包二进制滑窗大屏、TUN/TAP 虚拟网卡原始字节流接管、纯 Go 手写网络协议栈解码与履约核心。

Go 模块位于 `backend/`。评测入口：在该目录执行 `go test ./...`。

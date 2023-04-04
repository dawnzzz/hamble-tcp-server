package iface

import "net"

// IConnection 与客户端连接的抽象表示
type IConnection interface {
	Start()                // 启动连接，让当前连接开始工作
	Stop()                 // 停止连接，结束当前连接状态M
	GetConn() *net.TCPConn // 获取原始socket TCP连接
	RemoteAddr() string
}

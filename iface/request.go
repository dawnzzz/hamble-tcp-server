package iface

import "net"

// IRequest 客户端请求的抽象表示
type IRequest interface {
	GetConn() *net.TCPConn
	GetData() []byte
}

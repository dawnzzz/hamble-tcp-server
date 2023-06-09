package iface

// IRequest 客户端请求的抽象表示
type IRequest interface {
	GetConnection() IConnection
	GetData() []byte
	GetMsgID() uint32
}

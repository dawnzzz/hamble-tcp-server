package iface

// IServer TCP 服务器
type IServer interface {
	Start() // 开启服务器
	Stop()  // 结束服务器
	Serve() // 开始服务
}

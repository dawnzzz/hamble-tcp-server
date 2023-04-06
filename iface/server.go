package iface

import "time"

// IServer TCP 服务器
type IServer interface {
	ICSBase
	Start()                                      // 开启服务器
	Stop()                                       // 结束服务器
	Serve()                                      // 开始服务
	RegisterHandler(id uint32, handler IHandler) // 注册Handler
	StartHeartbeat(interval time.Duration)       // 开始心跳检测
	StartHeartbeatWithOption(CheckerOption)      // 开始心跳检测，使用CheckerOption
}

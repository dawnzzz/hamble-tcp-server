package iface

import "time"

type IClient interface {
	ICSBase
	Start()                                 // 开启客户端
	Stop()                                  // 结束客户端
	GetConnection() IConnection             // 获取连接
	StartHeartbeat(interval time.Duration)  // 开始心跳检测
	StartHeartbeatWithOption(CheckerOption) // 开始心跳检测，使用CheckerOption
}

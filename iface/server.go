package iface

import "time"

// IServer TCP 服务器
type IServer interface {
	Start()                                      // 开启服务器
	Stop()                                       // 结束服务器
	Serve()                                      // 开始服务
	RegisterHandler(id uint32, handler IHandler) // 注册Handler
	GetRouter() IRouter                          // 获取Router
	GetConnManager() IConnManager                // 获取ConnManager
	SetOnConnStart(func(conn IConnection))       // 设置连接创建时的Hook函数
	SetOnConnStop(func(conn IConnection))        // 设置连接结束时的Hook函数
	CallOnConnStart(conn IConnection)            // 调用连接创建时的Hook函数
	CallOnConnStop(conn IConnection)             // 调用连接结束时的Hook函数
	GetHeartBeatChecker() IHeartBeatChecker      // 获取心跳检测器
	StartHeartbeat(interval time.Duration)       // 开始心跳检测
}

package iface

// HeartBeatMsgFunc 用户自定义的心跳消息生成函数
type HeartBeatMsgFunc func(IConnection) []byte

// HeartBeatFunc 用户自定义心跳函数
type HeartBeatFunc func(IConnection) error

// OnRemoteNotAlive 远程连接不活跃时的处理方法
type OnRemoteNotAlive func(connection IConnection)

type IHeartBeatChecker interface {
	SetOnRemoteNotAlive(OnRemoteNotAlive)
	SetHeartbeatMsgFunc(HeartBeatMsgFunc)
	SetHeartbeatFunc(HeartBeatFunc)
	Start()
	Stop()
	SendHeartBeatMsg() error
	BindHandler(IHandler)
	BindConn(IConnection)
	Clone() IHeartBeatChecker
}

const DefaultHeartbeatMsgID = uint32(11111)

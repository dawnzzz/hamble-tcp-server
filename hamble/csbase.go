package hamble

import (
	"github.com/dawnzzz/hamble-tcp-server/hamble/heartbeat"
	"github.com/dawnzzz/hamble-tcp-server/iface"
	"github.com/dawnzzz/hamble-tcp-server/logger"
	"time"
)

// CSBase Server和Client的共同祖先，Server和Client都继承于此
type CSBase struct {
	dataPack iface.IDataPack // 封包解包方式
	router   iface.IRouter   // 路由模块

	connManager iface.IConnManager // 连接管理模块

	onConnStart func(connection iface.IConnection) // Hook
	onConnStop  func(connection iface.IConnection) // Hook

	checker iface.IHeartBeatChecker // 心跳检测
}

func (cs *CSBase) RegisterHandler(id uint32, handler iface.IHandler) {
	cs.router.AddRouter(id, handler)
}

func (cs *CSBase) GetRouter() iface.IRouter {
	return cs.router
}

func (cs *CSBase) GetConnManager() iface.IConnManager {
	return cs.connManager
}

func (cs *CSBase) SetOnConnStart(f func(conn iface.IConnection)) {
	cs.onConnStart = f
}

func (cs *CSBase) SetOnConnStop(f func(conn iface.IConnection)) {
	cs.onConnStop = f
}

func (cs *CSBase) CallOnConnStart(conn iface.IConnection) {
	if cs.onConnStart != nil {
		cs.onConnStart(conn)
	}
}

func (cs *CSBase) CallOnConnStop(conn iface.IConnection) {
	if cs.onConnStop != nil {
		cs.onConnStop(conn)
	}
}

func (cs *CSBase) StartHeartbeat(interval time.Duration) {
	cs.checker = heartbeat.NewHearBeatChecker(interval)
	cs.RegisterHandler(iface.DefaultHeartbeatMsgID, &heartbeat.DefaultHandler{})
}

func (cs *CSBase) StartHeartbeatWithOption(option iface.CheckerOption) {
	if option.Interval <= 0 {
		logger.Fatal("heartbeat checker interval must > 0")
		return
	}

	cs.checker = heartbeat.NewHearBeatChecker(option.Interval)

	if option.OnRemoteNotAlive != nil {
		cs.checker.SetOnRemoteNotAlive(option.OnRemoteNotAlive)
	}

	if option.HeartbeatMsgFunc != nil {
		cs.checker.SetOnRemoteNotAlive(option.OnRemoteNotAlive)
	}

	if option.Handler != nil {
		cs.RegisterHandler(option.MsgID, option.Handler)
	} else {
		cs.RegisterHandler(iface.DefaultHeartbeatMsgID, &heartbeat.DefaultHandler{})
	}
}

func (cs *CSBase) GetHeartBeatChecker() iface.IHeartBeatChecker {
	return cs.checker
}

func (cs *CSBase) GetDataPack() iface.IDataPack {
	return cs.dataPack
}

func (cs *CSBase) SetDataPack(dataPack iface.IDataPack) {
	if dataPack == nil {
		return
	}

	cs.dataPack = dataPack
}

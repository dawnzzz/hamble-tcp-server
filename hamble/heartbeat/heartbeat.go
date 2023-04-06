package heartbeat

import (
	"github.com/dawnzzz/hamble-tcp-server/iface"
	"github.com/dawnzzz/hamble-tcp-server/logger"
	"time"
)

// Checker  心跳检测器
type Checker struct {
	interval         time.Duration          // 心跳检查的时间间隔
	onRemoteNotAlive iface.OnRemoteNotAlive // 连接不存活时进行的处理
	heartbeatMsgFunc iface.HeartBeatMsgFunc // 用户自定义的心跳消息生成函数
	heartbeatFunc    iface.HeartBeatFunc    // 用户自定义心跳检测机制处理函数

	connection iface.IConnection // 心跳检测对应的连接
	msgID      uint32            // 心跳消息的消息id
	handler    iface.IHandler

	closedChan chan struct{}
}

func NewHearBeatChecker(interval time.Duration) iface.IHeartBeatChecker {
	checker := &Checker{
		interval:         interval,
		onRemoteNotAlive: defaultOnRemoteNotAlive,
		heartbeatMsgFunc: defaultHeartbeatMsgFunc,

		msgID:   iface.DefaultHeartbeatMsgID,
		handler: &DefaultHandler{},

		closedChan: make(chan struct{}, 1),
	}

	return checker
}

func (checker *Checker) SetOnRemoteNotAlive(onRemoteNotAlive iface.OnRemoteNotAlive) {
	checker.onRemoteNotAlive = onRemoteNotAlive
}

func (checker *Checker) SetHeartbeatMsgFunc(msgFunc iface.HeartBeatMsgFunc) {
	checker.heartbeatMsgFunc = msgFunc
}

func (checker *Checker) SetHeartbeatFunc(heartBeatFunc iface.HeartBeatFunc) {
	checker.heartbeatFunc = heartBeatFunc
}

func (checker *Checker) BindConn(connection iface.IConnection) {
	checker.connection = connection
}

func (checker *Checker) BindHandler(handler iface.IHandler) {
	checker.handler = handler
}

func (checker *Checker) Start() {
	go checker.start()
}

func (checker *Checker) start() {
	ticker := time.NewTicker(checker.interval)
	for {
		select {
		case <-ticker.C:
			_ = checker.check()
		case <-checker.closedChan:
			ticker.Stop()
			return
		}
	}
}

func (checker *Checker) check() error {
	// 首先检查连接是否存活
	if !checker.connection.IsAlive() {
		// 如果不存活
		checker.onRemoteNotAlive(checker.connection)
		checker.Stop()
		return nil
	}

	// 存活则发送心跳消息
	if checker.heartbeatFunc != nil {
		err := checker.heartbeatFunc(checker.connection)
		if err != nil {
			return err
		}
	} else {
		err := checker.SendHeartBeatMsg()
		if err != nil {
			return err
		}
	}

	return nil
}

func (checker *Checker) Stop() {
	checker.closedChan <- struct{}{}
}

func (checker *Checker) SendHeartBeatMsg() error {
	msg := checker.heartbeatMsgFunc(checker.connection)

	err := checker.connection.SendMsg(checker.msgID, msg)
	if err != nil {
		logger.Errorf("send heartbeat msg error: %v, msgId=%+v msg=%+v", err, checker.msgID, msg)
		return err
	}

	return nil
}

func (checker *Checker) Clone() iface.IHeartBeatChecker {
	return &Checker{
		interval:         checker.interval,
		onRemoteNotAlive: checker.onRemoteNotAlive,
		heartbeatMsgFunc: checker.heartbeatMsgFunc,
		heartbeatFunc:    checker.heartbeatFunc,
		connection:       nil,
		msgID:            checker.msgID,
		handler:          checker.handler,
		closedChan:       make(chan struct{}),
	}
}

package heartbeat

import (
	"fmt"
	"github.com/dawnzzz/hamble-tcp-server/iface"
	"github.com/dawnzzz/hamble-tcp-server/logger"
)

type DefaultHandler struct {
}

func (heartbeatHandler *DefaultHandler) PreHandle(_ iface.IRequest) {
}

func (heartbeatHandler *DefaultHandler) Handle(request iface.IRequest) {
	logger.Infof("receive heartbeat from %s, msgID=%v, data=%v", request.GetConnection().RemoteAddr(), request.GetMsgID(), request.GetData())
}

func (heartbeatHandler *DefaultHandler) PostHandle(_ iface.IRequest) {
}

func defaultHeartbeatMsgFunc(connection iface.IConnection) []byte {
	msg := fmt.Sprintf("heartbeat [%s->%s]", connection.GetConn().LocalAddr(), connection.GetConn().RemoteAddr())

	return []byte(msg)
}

func defaultOnRemoteNotAlive(connection iface.IConnection) {
	logger.Infof("connection from %s is not alive, stop it", connection.RemoteAddr())
	connection.Stop()
}

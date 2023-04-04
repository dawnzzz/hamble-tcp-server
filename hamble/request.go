package hamble

import (
	"github.com/dawnzzz/hamble-tcp-server/iface"
)

type Request struct {
	conn iface.IConnection
	data iface.IMessage
}

func NewRequest(conn iface.IConnection, message iface.IMessage) iface.IRequest {
	return &Request{
		conn: conn,
		data: message,
	}
}

func (req *Request) GetConnection() iface.IConnection {
	return req.conn
}

func (req *Request) GetData() []byte {
	return req.data.GetData()
}

func (req *Request) GetMsgID() uint32 {
	return req.data.GetMsgID()
}

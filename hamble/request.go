package hamble

import (
	"github.com/dawnzzz/hamble-tcp-server/iface"
	"net"
)

type Request struct {
	conn *net.TCPConn
	data iface.IMessage
}

func NewRequest(conn *net.TCPConn, message iface.IMessage) iface.IRequest {
	return &Request{
		conn: conn,
		data: message,
	}
}

func (req *Request) GetConn() *net.TCPConn {
	return req.conn
}

func (req *Request) GetData() []byte {
	return req.data.GetData()
}

func (req *Request) GetMsgID() uint32 {
	return req.data.GetMsgID()
}

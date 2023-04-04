package hamble

import (
	"github.com/dawnzzz/hamble-tcp-server/iface"
	"net"
)

type Request struct {
	conn *net.TCPConn
	data []byte
}

func NewRequest(conn *net.TCPConn, data []byte) iface.IRequest {
	return &Request{
		conn: conn,
		data: data,
	}
}

func (req *Request) GetConn() *net.TCPConn {
	return req.conn
}

func (req *Request) GetData() []byte {
	return req.data
}

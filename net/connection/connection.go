package connection

import (
	"github.com/dawnzzz/hamble-tcp-server/iface"
	"github.com/dawnzzz/hamble-tcp-server/logger"
	"net"
)

// Connection 与客户端的连接，实现了iface.IConnection接口
type Connection struct {
	conn *net.TCPConn // 原始 socket TCP 连接
}

func NewConnection(conn *net.TCPConn) iface.IConnection {
	return &Connection{
		conn: conn,
	}
}

func (c *Connection) startRead() {
	for {
		// todo:设计handler
		buf := make([]byte, 512)
		cnt, err := c.conn.Read(buf)
		if err != nil {
			c.Stop()
			return
		}

		logger.Infof("receive message from %s, message: %v", c.RemoteAddr(), buf[:cnt])
	}
}

func (c *Connection) Start() {
	logger.Infof("accept a connection from %s", c.RemoteAddr())

	go c.startRead()
}

func (c *Connection) Stop() {
	_ = c.conn.Close()
	logger.Infof("close a connection from %s", c.RemoteAddr())
}

func (c *Connection) GetConn() *net.TCPConn {
	return c.conn
}

func (c *Connection) RemoteAddr() string {
	return c.conn.RemoteAddr().String()
}

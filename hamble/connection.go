package hamble

import (
	"github.com/dawnzzz/hamble-tcp-server/iface"
	"github.com/dawnzzz/hamble-tcp-server/logger"
	"net"
	"sync"
)

// Connection 与客户端的连接，实现了iface.IConnection接口
type Connection struct {
	router iface.IRouter
	conn   *net.TCPConn // 原始 socket TCP 连接

	wg sync.WaitGroup
}

func NewConnection(conn *net.TCPConn, router iface.IRouter) iface.IConnection {
	return &Connection{
		conn:   conn,
		router: router,
	}
}

func (c *Connection) startRead() {
	for {
		buf := make([]byte, 512)
		cnt, err := c.conn.Read(buf)
		if err != nil {
			c.Stop()
			return
		}

		request := NewRequest(c.GetConn(), buf[:cnt])
		// 选择handler
		handler := c.router.GetHandler(0) // TODO:在请求中解析出ID
		go func() {
			c.wg.Add(1)
			defer c.wg.Done()
			handler.PreHandle(request)
			handler.Handle(request)
			handler.PostHandle(request)
		}()
	}
}

func (c *Connection) Start() {
	logger.Infof("accept a connection from %s", c.RemoteAddr())

	go c.startRead()

	c.wg.Wait()
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

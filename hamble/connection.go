package hamble

import (
	"github.com/dawnzzz/hamble-tcp-server/iface"
	"github.com/dawnzzz/hamble-tcp-server/logger"
	"io"
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
		dataPack := NewDataPack()

		buf := make([]byte, dataPack.GetHeadLen())
		_, err := c.conn.Read(buf)
		if err != nil {
			c.Stop()
			return
		}

		// 解包
		msg, err := dataPack.Unpack(buf)
		if err != nil {
			c.Stop()
			return
		}
		dataBuf := make([]byte, msg.GetDataLen())
		_, err = io.ReadFull(c.conn, dataBuf)
		if err != nil {
			c.Stop()
			return
		}
		msg.SetData(dataBuf)

		request := NewRequest(c.GetConn(), msg)
		// 选择handler
		handler := c.router.GetHandler(msg.GetMsgID())
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

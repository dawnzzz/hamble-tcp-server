package hamble

import (
	"errors"
	"github.com/dawnzzz/hamble-tcp-server/iface"
	"github.com/dawnzzz/hamble-tcp-server/logger"
	"io"
	"net"
	"sync"
	"sync/atomic"
)

// Connection 与客户端的连接，实现了iface.IConnection接口
type Connection struct {
	router iface.IRouter
	conn   *net.TCPConn // 原始 socket TCP 连接

	wg       sync.WaitGroup
	isClosed atomic.Bool
}

func NewConnection(conn *net.TCPConn, router iface.IRouter) iface.IConnection {
	return &Connection{
		conn:   conn,
		router: router,
	}
}

func (c *Connection) startRead() {
	for {
		if c.isClosed.Load() {
			return
		}

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

		request := NewRequest(c, msg)
		go func() {
			c.wg.Add(1)
			defer c.wg.Done()
			c.router.DoHandler(request) // 执行 handler
		}()
	}
}

func (c *Connection) Start() {
	logger.Infof("accept a connection from %s", c.RemoteAddr())

	go c.startRead()

	c.wg.Wait()
}

func (c *Connection) Stop() {
	c.isClosed.Store(true)
	_ = c.conn.Close()

	logger.Infof("close a connection from %s", c.RemoteAddr())
}

func (c *Connection) GetConn() *net.TCPConn {
	return c.conn
}

func (c *Connection) RemoteAddr() string {
	return c.conn.RemoteAddr().String()
}

func (c *Connection) SendMsg(msgID uint32, data []byte) error {
	if c.isClosed.Load() {
		// 关闭直接返回
		return errors.New("connection closed when send msg")
	}

	// 将消息封包，发送
	dp := NewDataPack()
	msg := NewMessage(msgID, data)
	packet, err := dp.Pack(msg)
	if err != nil {
		return err
	}

	if _, err = c.conn.Write(packet); err != nil {
		// 发送失败，关闭连接
		c.Stop()

		return err
	}

	return nil
}

package hamble

import (
	"errors"
	"github.com/dawnzzz/hamble-tcp-server/conf"
	"github.com/dawnzzz/hamble-tcp-server/iface"
	"github.com/dawnzzz/hamble-tcp-server/logger"
	"io"
	"net"
	"sync/atomic"
)

// Connection 与客户端的连接，实现了iface.IConnection接口
type Connection struct {
	router iface.IRouter
	conn   *net.TCPConn // 原始 socket TCP 连接

	msgChan chan iface.IMessage // 服务器待发送的消息放在这里

	exitChan chan struct{}
	isClosed atomic.Bool
}

func NewConnection(conn *net.TCPConn, router iface.IRouter) iface.IConnection {
	return &Connection{
		conn:   conn,
		router: router,

		msgChan:  make(chan iface.IMessage),
		exitChan: make(chan struct{}, 1),
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
			c.exitChan <- struct{}{}
			return
		}

		// 解包
		msg, err := dataPack.Unpack(buf)
		if err != nil {
			c.exitChan <- struct{}{}
			return
		}
		dataBuf := make([]byte, msg.GetDataLen())
		_, err = io.ReadFull(c.conn, dataBuf)
		if err != nil {
			c.exitChan <- struct{}{}
			return
		}
		msg.SetData(dataBuf)

		request := NewRequest(c, msg)

		if conf.GlobalProfile.WorkerPoolSize > 0 {
			//已经启动工作池机制，将消息交给Worker处理
			c.router.SendMsgToTaskQueue(request)
		} else {
			go func() {
				c.router.DoHandler(request) // 执行 handler
			}()
		}

	}
}

func (c *Connection) startWrite() {
	// 将消息封包，发送
	dp := NewDataPack()

	for msg := range c.msgChan {
		packet, err := dp.Pack(msg)
		if err != nil {
			c.exitChan <- struct{}{}
			return
		}

		if _, err = c.conn.Write(packet); err != nil {
			// 发送失败，关闭连接
			c.exitChan <- struct{}{}
			return
		}
	}
}

func (c *Connection) Start() {
	logger.Infof("accept a connection from %s", c.RemoteAddr())

	go c.startRead()
	go c.startWrite()

	// 阻塞，直到退出
	select {
	case <-c.exitChan:
		return
	}
}

func (c *Connection) Stop() {
	if !c.isClosed.CompareAndSwap(false, true) {
		// 已经关闭，直接返回
		return
	}

	// 关闭管道
	close(c.msgChan)
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
		c.exitChan <- struct{}{}
		return errors.New("connection closed when send msg")
	}

	msg := NewMessage(msgID, data)

	// 将消息推入c.msgChan，等待发送
	c.msgChan <- msg

	return nil
}

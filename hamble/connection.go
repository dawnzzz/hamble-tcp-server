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
	TcpServer iface.IServer

	conn *net.TCPConn // 原始 socket TCP 连接

	msgChan    chan iface.IMessage // 服务器待发送的消息放在这里
	msgBufChan chan iface.IMessage // 带缓冲区的msgChan

	exitChan chan struct{}
	isClosed atomic.Bool
}

func NewConnection(conn *net.TCPConn, server iface.IServer) iface.IConnection {
	return &Connection{
		TcpServer: server,
		conn:      conn,

		msgChan:    make(chan iface.IMessage),
		msgBufChan: make(chan iface.IMessage, conf.GlobalProfile.MaxMsgChanLen),
		exitChan:   make(chan struct{}, 1),
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
			c.TcpServer.GetRouter().SendMsgToTaskQueue(request)
		} else {
			go func() {
				c.TcpServer.GetRouter().DoHandler(request) // 执行 handler
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

func (c *Connection) startBufWrite() {
	// 将消息封包，发送
	dp := NewDataPack()

	for msg := range c.msgBufChan {
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

	c.TcpServer.GetConnManager().Add(c)

	go c.startRead()
	go c.startWrite()
	go c.startBufWrite()

	c.TcpServer.GetConnManager().Add(c) // 将当前连接添加到连接管理器中

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
	close(c.msgBufChan)
	_ = c.conn.Close()
	c.TcpServer.GetConnManager().Remove(c)

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

func (c *Connection) SendBufMsg(msgID uint32, data []byte) error {
	if c.isClosed.Load() {
		// 关闭直接返回
		c.exitChan <- struct{}{}
		return errors.New("connection closed when send msg")
	}

	msg := NewMessage(msgID, data)

	// 将消息推入c.msgChan，等待发送
	c.msgBufChan <- msg

	return nil
}

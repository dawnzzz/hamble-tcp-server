package hamble

import (
	"errors"
	"github.com/dawnzzz/hamble-tcp-server/conf"
	"github.com/dawnzzz/hamble-tcp-server/iface"
	"github.com/dawnzzz/hamble-tcp-server/logger"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// Connection 与客户端的连接，实现了iface.IConnection接口
type Connection struct {
	cs iface.ICSBase // 指向客户端或者服务器（client/server）

	conn net.Conn // 原始 socket TCP 连接

	msgChan    chan iface.IMessage // 服务器待发送的消息放在这里
	msgBufChan chan iface.IMessage // 带缓冲区的msgChan

	exitChan chan struct{}
	isClosed atomic.Bool

	properties     map[string]interface{} //	记录连接属性
	propertiesLock sync.Mutex             // 保证连接属性的互斥访问

	heartbeatChecker iface.IHeartBeatChecker
	lastAliveTime    time.Time // 上一次收到消息的时间，用于在心跳检测中检查是否存活
}

func newConnection(conn net.Conn, cs iface.ICSBase) iface.IConnection {
	return &Connection{
		cs:   cs,
		conn: conn,

		msgChan:    make(chan iface.IMessage, 1),
		msgBufChan: make(chan iface.IMessage, conf.GlobalProfile.MaxMsgChanLen),
		exitChan:   make(chan struct{}, 1),

		lastAliveTime: time.Now(),
	}
}

func (c *Connection) updateLastAliveTime(newTime time.Time) {
	c.lastAliveTime = newTime
}

func (c *Connection) startRead() {
	for {
		if c.isClosed.Load() {
			return
		}

		dataPack := c.cs.GetDataPack()

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

		// 记录收到消息的时间
		c.updateLastAliveTime(time.Now())

		request := NewRequest(c, msg)

		if conf.GlobalProfile.WorkerPoolSize > 0 {
			//已经启动工作池机制，将消息交给Worker处理
			c.cs.GetRouter().SendMsgToTaskQueue(request)
		} else {
			go func() {
				c.cs.GetRouter().DoHandler(request) // 执行 handler
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

	c.cs.GetConnManager().Add(c)

	go c.startRead()
	go c.startWrite()
	go c.startBufWrite()

	if c.cs.GetHeartBeatChecker() != nil {
		// 开启心跳检测
		heartbeatChecker := c.cs.GetHeartBeatChecker().Clone()
		heartbeatChecker.BindConn(c)
		c.heartbeatChecker = heartbeatChecker
		c.heartbeatChecker.Start()
	}

	c.cs.GetConnManager().Add(c) // 将当前连接添加到连接管理器中

	// 执行Hook函数
	c.cs.CallOnConnStart(c)

	// 阻塞，直到退出
	select {
	case <-c.exitChan:
		c.Stop()
		return
	}
}

func (c *Connection) Stop() {
	if !c.isClosed.CompareAndSwap(false, true) {
		// 已经关闭，直接返回
		return
	}

	// 执行Hook函数
	c.cs.CallOnConnStop(c)

	// 关闭管道
	close(c.msgChan)
	close(c.msgBufChan)
	_ = c.conn.Close()
	c.cs.GetConnManager().Remove(c)
	if c.heartbeatChecker != nil {
		c.heartbeatChecker.Stop()
	}

	logger.Infof("close a connection from %s", c.RemoteAddr())
}

func (c *Connection) GetConn() net.Conn {
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

func (c *Connection) SetProperty(key string, value interface{}) {
	c.propertiesLock.Lock()
	defer c.propertiesLock.Unlock()

	if c.properties == nil {
		c.properties = make(map[string]interface{}) // 延迟初始化
	}

	c.properties[key] = value
}

func (c *Connection) GetProperty(key string) interface{} {
	c.propertiesLock.Lock()
	defer c.propertiesLock.Unlock()

	if c.properties == nil {
		return nil
	}

	value, exist := c.properties[key]
	if !exist {
		return nil
	}

	return value
}

func (c *Connection) RemoveProperty(key string) {
	c.propertiesLock.Lock()
	defer c.propertiesLock.Unlock()

	if c.properties == nil {
		return
	}

	delete(c.properties, key)
}

func (c *Connection) IsAlive() bool {
	if c.isClosed.Load() {
		// 连接已经关闭
		return false
	}

	return conf.GlobalProfile.MaxHeartbeatTime <= 0 || time.Now().Before(c.lastAliveTime.Add(conf.GlobalProfile.GetMaxHeartbeatTime()))
}

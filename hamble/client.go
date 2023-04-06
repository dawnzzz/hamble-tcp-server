package hamble

import (
	"fmt"
	"github.com/dawnzzz/hamble-tcp-server/conf"
	"github.com/dawnzzz/hamble-tcp-server/iface"
	"github.com/dawnzzz/hamble-tcp-server/logger"
	"net"
	"sync"
)

// Client 客户端
type Client struct {
	CSBase

	Version    string // TCP版本号 tcp or tcp4 or tcp6
	IP         string // 客户端连接地址
	Port       int    // 客户端连接端口号
	connection iface.IConnection

	exitChan chan struct{}
	wg       sync.WaitGroup
}

func NewClient(network string, ip string, port int) (iface.IClient, error) {

	router := newRouter()

	c := &Client{
		CSBase: CSBase{
			router:      router,
			dataPack:    NewDataPack(),
			connManager: NewConnManager(),
		},
		Version:    network,
		IP:         ip,
		Port:       port,
		connection: nil,
		exitChan:   make(chan struct{}, 1),
	}

	// 发起连接
	addr, err := net.ResolveTCPAddr(network, fmt.Sprintf("%v:%v", ip, port))
	if err != nil {
		logger.Errorf("resolve tcp addr err: %s", err.Error())
		return nil, err
	}

	conn, err := net.DialTCP(network, nil, addr)
	if err != nil {
		logger.Errorf("dial tcp err: %s", err.Error())
		return nil, err
	}

	// 创建新的连接
	c.connection = newConnection(conn, c)

	return c, nil
}

func (c *Client) Start() {
	//客户端将协程池关闭
	conf.GlobalProfile.WorkerPoolSize = 0

	logger.Infof("client start")

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()

		// 启动连接
		c.connection.Start()

		// 启动一个线程检查退出信号
		go func() {
			select {
			case <-c.exitChan:
				// 停止连接
				c.connection.Stop()
			}
		}()
	}()

	// 等待结束
	c.wg.Wait()
}

func (c *Client) Stop() {
	c.exitChan <- struct{}{}
	logger.Infof("client stop")
}

func (c *Client) GetConnection() iface.IConnection {
	return c.connection
}

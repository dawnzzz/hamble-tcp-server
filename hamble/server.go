package hamble

import (
	"context"
	"errors"
	"fmt"
	"github.com/dawnzzz/hamble-tcp-server/conf"
	"github.com/dawnzzz/hamble-tcp-server/iface"
	"github.com/dawnzzz/hamble-tcp-server/logger"
	"github.com/sirupsen/logrus"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// Server TCP服务器，实现了iface.IServer接口
type Server struct {
	Name    string // 服务器名称
	Version string // TCP版本号 tcp or tcp4 or tcp6
	IP      string // 服务器监听地址
	Port    int    // 服务器端口号

	router iface.IRouter

	connManager iface.IConnManager // 连接管理模块

	ctx         context.Context
	cancel      context.CancelFunc // 提醒Server退出
	wg          sync.WaitGroup
	closingChan chan struct{} // 发送退出信号
}

func NewServer() iface.IServer {
	conf.Reload() // 加载配置文件

	ctx, cancel := context.WithCancel(context.Background())

	router := newRouter()

	s := &Server{
		router: router,

		Name:    conf.GlobalProfile.Name,
		Version: conf.GlobalProfile.TcpVersion,
		IP:      conf.GlobalProfile.Host,
		Port:    conf.GlobalProfile.Port,

		connManager: NewConnManager(),

		ctx:         ctx,
		cancel:      cancel,
		closingChan: make(chan struct{}, 1),
	}

	logger.WithFields(logrus.Fields{
		"TCPServer": "Hamble",
		"Name":      s.Name,
	})

	return s
}

// Start 开启hamble TCP 服务器，当调用此函数时，当前协程会阻塞住进行TCP服务
func (s *Server) Start() {
	logger.Infof("server start")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	// 开启一个协程监听退出信号
	go func() {
		select {
		case <-sigChan:
			// 收到退出信号，向closingChan中发送消息，表示需要退出TCP服务器
			s.closingChan <- struct{}{}
			close(s.closingChan)
			return
		}
	}()

	// 开启一个协程监听closingChan，若收到消息则退出TCP服务器
	go func() {
		select {
		case <-s.closingChan:
			// 调用 s.Stop 结束服务器
			s.Stop()
			return
		}
	}()

	// 开启工作池
	s.router.StartWorkerPool()

	// 开启服务
	s.wg.Add(1)
	go s.Serve()

	// 等待 s.Server 退出
	s.wg.Wait()

	logger.Info("server exited")
}

// Stop 停止 TCP 服务器
func (s *Server) Stop() {
	logger.Infof("server stop")

	// 退出之前关闭全部连接，实现优雅的关闭
	s.connManager.Clear()

	// 调用cancel取消
	s.cancel()
}

func (s *Server) Serve() {
	defer s.wg.Done() // 通知主线程退出

	// 开始正常的服务
	addr, err := net.ResolveTCPAddr(s.Version, fmt.Sprintf("%s:%d", s.IP, s.Port))
	if err != nil {
		logger.Error("resolve tcp addr err,", err)
		return
	}

	listener, err := net.ListenTCP(s.Version, addr)
	if err != nil {
		logger.Error("listen tcp err,", err)
		return
	}

	// 开启一个协程检查退出信号
	go func() {
		select {
		case <-s.ctx.Done():
			// 需要退出服务器了
			_ = listener.Close() // 关闭 listener
			return
		}
	}()

	logger.Info("start listen")

	for {

		// 等待accept
		tcpConn, err := listener.AcceptTCP()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				// 连接关闭，直接退出
				return
			}
			// 遇到其他错误跳过
			logger.Info("accept tcp err, ", err)
			continue
		}

		if s.connManager.Len() >= conf.GlobalProfile.MaxConn {
			// 超过了最大连接数，直接关闭连接
			_ = tcpConn.Close()
			continue
		}

		conn := NewConnection(tcpConn, s)
		go func() {
			defer func() {
				conn.Stop()
			}()
			conn.Start() // 连接开始工作
		}()
	}
}

func (s *Server) RegisterHandler(id uint32, handler iface.IHandler) {
	s.router.AddRouter(id, handler)
}

func (s *Server) GetRouter() iface.IRouter {
	return s.router
}

func (s *Server) GetConnManager() iface.IConnManager {
	return s.connManager
}

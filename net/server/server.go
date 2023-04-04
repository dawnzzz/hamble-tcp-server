package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/dawnzzz/hamble-tcp-server/iface"
	"github.com/dawnzzz/hamble-tcp-server/logger"
	"github.com/dawnzzz/hamble-tcp-server/net/connection"
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

	connections map[iface.IConnection]struct{} // 记录当前的连接

	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	closingChan chan struct{} // 发送退出信号
}

func NewServer() iface.IServer {
	ctx, cancel := context.WithCancel(context.Background())

	s := &Server{
		Name:    "hamble-tcp-server",
		Version: "tcp",
		IP:      "0.0.0.0",
		Port:    6177,

		connections: make(map[iface.IConnection]struct{}),

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
	logger.Infof("started")

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

	// 开启服务
	s.wg.Add(1)
	go s.Serve()

	// 等待 s.Server 退出
	s.wg.Wait()

	logger.Info("exited")
}

// Stop 停止 TCP 服务器
func (s *Server) Stop() {
	logger.Infof("stop")

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

	defer func() {
		// 退出之前关闭全部连接，实现优雅的关闭
		for conn := range s.connections {
			conn.Stop()
		}

		logger.Infof("serve is closed")
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

		conn := connection.NewConnection(tcpConn)
		s.connections[conn] = struct{}{}
		go conn.Start() // 连接开始工作
	}
}

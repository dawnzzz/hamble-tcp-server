package hamble

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/dawnzzz/hamble-tcp-server/conf"
	"github.com/dawnzzz/hamble-tcp-server/hamble/heartbeat"
	"github.com/dawnzzz/hamble-tcp-server/iface"
	"github.com/dawnzzz/hamble-tcp-server/logger"
	"github.com/dawnzzz/hamble-tcp-server/utils"
	"github.com/sirupsen/logrus"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Server TCP服务器，实现了iface.IServer接口
type Server struct {
	CSBase
	Name    string // 服务器名称
	Version string // TCP版本号 tcp or tcp4 or tcp6
	IP      string // 服务器监听地址
	Port    int    // 服务器端口号

	ctx         context.Context
	cancel      context.CancelFunc // 提醒Server退出
	wg          sync.WaitGroup
	closingChan chan struct{} // 发送退出信号

	useTLS bool
}

func NewServer() iface.IServer {
	conf.Reload() // 加载配置文件

	ctx, cancel := context.WithCancel(context.Background())

	router := newRouter()

	s := &Server{
		CSBase: CSBase{
			router:      router,
			dataPack:    NewDataPack(),
			connManager: NewConnManager(),
		},

		Name:    conf.GlobalProfile.Name,
		Version: conf.GlobalProfile.TcpVersion,
		IP:      conf.GlobalProfile.Host,
		Port:    conf.GlobalProfile.Port,

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

func NewTLSServer() iface.IServer {
	iServer := NewServer()
	s, _ := iServer.(*Server)
	s.useTLS = true

	return s
}

func NewServerWithOption(option *conf.Profile) iface.IServer {
	conf.BindProfile(option)

	ctx, cancel := context.WithCancel(context.Background())

	router := newRouter()

	s := &Server{
		CSBase: CSBase{
			router:      router,
			dataPack:    NewDataPack(),
			connManager: NewConnManager(),
		},

		Name:    conf.GlobalProfile.Name,
		Version: conf.GlobalProfile.TcpVersion,
		IP:      conf.GlobalProfile.Host,
		Port:    conf.GlobalProfile.Port,

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

func NewTLSServerWithOption(option *conf.Profile) iface.IServer {
	iServer := NewServerWithOption(option)

	s, _ := iServer.(*Server)
	s.useTLS = true

	return s
}

const banner = `
 ___  ___  ________  _____ ______   ________  ___       _______      
|\  \|\  \|\   __  \|\   _ \  _   \|\   __  \|\  \     |\  ___ \     
\ \  \\\  \ \  \|\  \ \  \\\__\ \  \ \  \|\ /\ \  \    \ \   __/|    
 \ \   __  \ \   __  \ \  \\|__| \  \ \   __  \ \  \    \ \  \_|/__  
  \ \  \ \  \ \  \ \  \ \  \    \ \  \ \  \|\  \ \  \____\ \  \_|\ \ 
   \ \__\ \__\ \__\ \__\ \__\    \ \__\ \_______\ \_______\ \_______\
    \|__|\|__|\|__|\|__|\|__|     \|__|\|_______|\|_______|\|_______|`

const url = "https://github.com/dawnzzz/hamble-tcp-server"

// Start 开启hamble TCP 服务器，当调用此函数时，当前协程会阻塞住进行TCP服务
func (s *Server) Start() {
	if conf.GlobalProfile.LogFileName != "" {
		logFile, err := os.OpenFile(conf.GlobalProfile.LogFileName, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
		if err != nil {
			logger.Error("open log file [%s] err: %v", conf.GlobalProfile.LogFileName, err.Error())
		} else {
			logger.SetMultiOutPut(logFile)
		}
	}

	if conf.GlobalProfile.PrintBanner {
		fmt.Printf("%s\n\npowered by %s\n\n", banner, url)
	}

	conf.PrintGlobalProfile()

	if s.useTLS {
		logger.Info("!! Attention, Server is using TLS !! ")
	}

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

	var listener net.Listener
	if s.useTLS {
		// 使用 TLS 加密

		// 必要时生成私钥和证书文件
		if !utils.IsFileExist(conf.GlobalProfile.CrtFileName) && !utils.IsFileExist(conf.GlobalProfile.KeyFileName) {
			err := utils.GenerateCrtAndKeyFile(conf.GlobalProfile.CrtFileName, conf.GlobalProfile.KeyFileName)
			if err != nil {
				logger.Errorf("create crt and private key err: %v", err.Error())
				return
			}
		}

		if !utils.IsFileExist(conf.GlobalProfile.CrtFileName) || !utils.IsFileExist(conf.GlobalProfile.KeyFileName) {
			logger.Error("CRT file and PrivateKey File must both exist or both not exist!!!")
		}

		// 读取证书和密钥
		crt, err := tls.LoadX509KeyPair(conf.GlobalProfile.CrtFileName, conf.GlobalProfile.KeyFileName)
		if err != nil {
			logger.Errorf("load x509 err: %v", err.Error())
			return
		}

		// TLS连接
		tlsConfig := &tls.Config{}
		tlsConfig.Certificates = []tls.Certificate{crt}
		tlsConfig.Time = time.Now
		tlsConfig.Rand = rand.Reader
		listener, err = tls.Listen(s.Version, fmt.Sprintf("%s:%d", s.IP, s.Port), tlsConfig)
		if err != nil {
			logger.Errorf("listen tcp tls err: %v", err)
			return
		}
	} else {
		// 开始正常的服务
		addr, err := net.ResolveTCPAddr(s.Version, fmt.Sprintf("%s:%d", s.IP, s.Port))
		if err != nil {
			logger.Error("resolve tcp addr err,", err)
			return
		}

		listener, err = net.ListenTCP(s.Version, addr)
		if err != nil {
			logger.Error("listen tcp err,", err)
			return
		}
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
		tcpConn, err := listener.Accept()
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

		conn := newConnection(tcpConn, s)
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

func (s *Server) SetOnConnStart(f func(conn iface.IConnection)) {
	s.onConnStart = f
}

func (s *Server) SetOnConnStop(f func(conn iface.IConnection)) {
	s.onConnStop = f
}

func (s *Server) CallOnConnStart(conn iface.IConnection) {
	if s.onConnStart != nil {
		s.onConnStart(conn)
	}
}

func (s *Server) CallOnConnStop(conn iface.IConnection) {
	if s.onConnStop != nil {
		s.onConnStop(conn)
	}
}

func (s *Server) StartHeartbeat(interval time.Duration) {
	s.checker = heartbeat.NewHearBeatChecker(interval)
	s.RegisterHandler(iface.DefaultHeartbeatMsgID, &heartbeat.DefaultHandler{})
}

func (s *Server) StartHeartbeatWithOption(option iface.CheckerOption) {
	if option.Interval <= 0 {
		logger.Fatal("heartbeat checker interval must > 0")
		return
	}

	s.checker = heartbeat.NewHearBeatChecker(option.Interval)

	if option.OnRemoteNotAlive != nil {
		s.checker.SetOnRemoteNotAlive(option.OnRemoteNotAlive)
	}

	if option.HeartbeatMsgFunc != nil {
		s.checker.SetOnRemoteNotAlive(option.OnRemoteNotAlive)
	}

	if option.Handler != nil {
		s.RegisterHandler(option.MsgID, option.Handler)
	} else {
		s.RegisterHandler(iface.DefaultHeartbeatMsgID, &heartbeat.DefaultHandler{})
	}
}

func (s *Server) GetHeartBeatChecker() iface.IHeartBeatChecker {
	return s.checker
}

func (s *Server) GetDataPack() iface.IDataPack {
	return s.dataPack
}

func (s *Server) SetDataPack(dataPack iface.IDataPack) {
	if dataPack == nil {
		return
	}

	s.dataPack = dataPack
}

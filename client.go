package main

import (
	"fmt"
	"github.com/dawnzzz/hamble-tcp-server/hamble"
	"github.com/dawnzzz/hamble-tcp-server/iface"
	"time"
)

type PongHandler struct {
	hamble.BaseHandler
}

func (ph *PongHandler) Handle(request iface.IRequest) {
	fmt.Printf("resevied request: %s\n", request.GetData())
}

type RecvEchoHandler struct {
	hamble.BaseHandler
}

func (eh *RecvEchoHandler) Handle(request iface.IRequest) {
	fmt.Printf("resevied request: %s\n", request.GetData())
}

func main() {
	c, err := hamble.NewTLSClient("tcp", "127.0.0.1", 6177)
	if err != nil {
		fmt.Println(err)
		return
	}

	c.RegisterHandler(0, &PongHandler{})
	c.RegisterHandler(1, &RecvEchoHandler{})
	c.StartHeartbeat(5 * time.Second)

	// 设置hook函数
	c.SetOnConnStart(func(conn iface.IConnection) {
		fmt.Printf("on conn start\n")
	})
	c.SetOnConnStop(func(conn iface.IConnection) {
		fmt.Printf("on conn stop\n")
	})

	// 启动客户端
	go c.Start()

	// 发送消息
	for i := 0; i < 10; i++ {
		if i%2 == 0 {
			err = c.GetConnection().SendMsg(0, []byte("PING"))
		} else {
			err = c.GetConnection().SendBufMsg(1, []byte("Hello"))
		}

		if err != nil {
			c.GetConnection().Stop()
			return
		}

		time.Sleep(time.Second)
	}

	// 停止
	c.Stop()

}

package main

import (
	"fmt"
	"github.com/dawnzzz/hamble-tcp-server/hamble"
	"github.com/dawnzzz/hamble-tcp-server/iface"
	"time"
)

type PingHandler struct {
	hamble.BaseHandler
}

func (ph *PingHandler) Handle(request iface.IRequest) {
	fmt.Printf("resevied request: %s\n", request.GetData())
	_ = request.GetConnection().SendBufMsg(request.GetMsgID(), []byte("PONG"))
}

type EchoHandler struct {
	hamble.BaseHandler
}

func (eh *EchoHandler) Handle(request iface.IRequest) {
	fmt.Printf("resevied request: %s\n", request.GetData())
	_ = request.GetConnection().SendMsg(request.GetMsgID(), request.GetData())
}

func main() {
	s := hamble.NewTLSServer()
	s.RegisterHandler(0, &PingHandler{})
	s.RegisterHandler(1, &EchoHandler{})
	s.StartHeartbeat(5 * time.Second)
	s.Start()

}

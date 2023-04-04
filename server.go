package main

import (
	"fmt"
	"github.com/dawnzzz/hamble-tcp-server/hamble"
	"github.com/dawnzzz/hamble-tcp-server/iface"
)

type PingHandler struct {
	hamble.BaseHandler
}

func (ph *PingHandler) Handle(request iface.IRequest) {
	fmt.Printf("resevied request: %s\n", request.GetData())
	_, _ = request.GetConn().Write([]byte("PONG"))
}

func main() {
	s := hamble.NewServer()
	s.RegisterHandler(0, &PingHandler{})
	s.Start()
}

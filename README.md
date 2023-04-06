# Hamble

Hamble 是一个 golang 编写的 TCP 服务器。

## 使用方法

### 服务器 server

```go
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
   s := hamble.NewServer()
   s.RegisterHandler(0, &PingHandler{})
   s.RegisterHandler(1, &EchoHandler{})
   s.StartHeartbeat(5 * time.Second)
   s.Start()

}
```

### 客户端 client

```go
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
   c, err := hamble.NewClient("tcp", "127.0.0.1", 6177)
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
```
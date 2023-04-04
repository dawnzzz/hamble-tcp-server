package main

import (
	"fmt"
	"github.com/dawnzzz/hamble-tcp-server/hamble"
	"github.com/dawnzzz/hamble-tcp-server/iface"
	"net"
	"time"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:6177")
	if err != nil {
		println(err)
		return
	}

	dp := hamble.NewDataPack()
	for i := 0; i < 10; i++ {
		var msg iface.IMessage

		if i%2 == 0 {
			msg = hamble.NewMessage(0, []byte("PING"))
		} else {
			msg = hamble.NewMessage(1, []byte("Hello"))
		}

		data, err := dp.Pack(msg)
		if err != nil {
			return
		}

		_, err = conn.Write(data)
		if err != nil {
			return
		}

		buf := make([]byte, 512)
		cnt, err := conn.Read(buf)
		if err != nil {
			return
		}

		fmt.Printf("receive from server: %s\n", buf[:cnt])
		time.Sleep(time.Second)
	}

	_ = conn.Close()

}

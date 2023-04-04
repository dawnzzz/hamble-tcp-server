package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:6177")
	if err != nil {
		println(err)
		return
	}

	for i := 0; i < 10; i++ {
		_, err = conn.Write([]byte("PING "))
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

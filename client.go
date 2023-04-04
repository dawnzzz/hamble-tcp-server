package main

import (
	"net"
	"strconv"
	"time"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:6177")
	if err != nil {
		println(err)
		return
	}

	for i := 0; i < 10; i++ {
		conn.Write([]byte("hello " + strconv.Itoa(i)))
		time.Sleep(time.Second)
	}

	conn.Close()

}

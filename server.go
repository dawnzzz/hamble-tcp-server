package main

import "github.com/dawnzzz/hamble-tcp-server/net/server"

func main() {
	s := server.NewServer()
	s.Start()
}

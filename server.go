package main

import (
	"github.com/dawnzzz/hamble-tcp-server/hamble"
)

func main() {
	s := hamble.NewServer()
	s.Start()
}

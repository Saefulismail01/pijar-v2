package main

import (
	"pijar/delivery"
)

func main() {
	server := delivery.NewServer()
	server.Run()
}

package main

import (
	"fmt"
	"pijar/delivery"
)

func main() {
	server := delivery.NewServer()
	if server == nil {
		fmt.Println("Failed to initialize server. Please check your configuration and try again.")
		return
	}
	server.Run()
}

// package main
//
// import (
// 	"fmt"
// 	"pijar/delivery"
// )
//
// func main() {
// 	server := delivery.NewServer()
// 	if server == nil {
// 		fmt.Println("Failed to initialize server. Please check your configuration and try again.")
// 		return
// 	}
// 	server.Run()
// }


package main

import (
	"log"
	"pijar/delivery"
	_ "pijar/docs"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file")
	}

	delivery.NewServer().Run()
}

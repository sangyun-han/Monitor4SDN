package main

import (
	controller "github.com/sangyun-han/monitor4sdn/controller"
	"log"
	"os"
)

var logger *log.Logger

func main() {
	logger = log.New(os.Stdout, "[INFO][MAIN] ", log.LstdFlags)
	logger.Println("Start controller")

	ofc := controller.NewOFController()

	// start server
	ofc.Listen(controller.DEFAULT_PORT)

}

package main

import (
	controller "github.com/sangyun-han/monitor4sdn/controller"
	"log"
	"os"
)

var logger *log.Logger

type SampleController struct {
	// add any paramter used in controller.
}

func NewSampleController() *SampleController {
	ofc := new(SampleController)
	return ofc
}


func main() {
	logger = log.New(os.Stdout, "[INFO][MAIN] ", log.LstdFlags)
	logger.Println("Start controller")

	ofc := controller.NewOFController()
	controller.GetAppManager().RegisterApplication(ofc)

	// start server
	controller.ServerLoop(controller.DEFAULT_PORT)

}

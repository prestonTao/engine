package main

import (
	// "fmt"
	"github.com/prestonTao/engine"
	"time"
)

func main() {
	example1()
}

func example1() {
	engine.GlobalInit("console", "", "debug", 1)
	// server := engine.NewNet("tao")
	engine.RegisterMsg(101, hello)
	// server

	engine.InitEngine("file_server")
	// engine.SetAuth(new(handlers.NoneAuth))
	// engine.SetCloseCallback(handlers.CloseConnHook)
	err := engine.Listen("127.0.0.1", int32(9981))
	if err != nil {
		panic(err)
	}
	time.Sleep(time.Minute * 10)
}

func hello(c engine.Controller, msg engine.Packet) {

}

type FindNode struct {
	Name string `json:"name"`
}

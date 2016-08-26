package main

import (
	"fmt"
	"time"

	"github.com/prestonTao/engine"
)

func main() {
	example1()
}

var CryKey = []byte{0, 0, 0, 0, 0, 0, 0, 0}

func example1() {
	engine.GlobalInit("console", "", "debug", 1)

	engine.InitEngine("client_server")
	engine.RegisterMsg(101, hello)

	for {
		for j := 0; j < 100; j++ {
			go func() {
				for i := 0; i < 100000000000; i++ {
					nameOne, err := engine.AddClientConn("127.0.0.1", int32(9981), false)
					if err != nil {
						fmt.Println(err.Error())
					}
					session, ok := engine.GetSession(nameOne)
					if ok {
						data := []byte("hello")
						session.Send(101, 0, 0, CryKey, &data)
						session.Close()
					}
				}
			}()
		}
		time.Sleep(time.Minute)
	}

	time.Sleep(time.Hour * 100000)
}

func hello(c engine.Controller, msg engine.Packet) {
	fmt.Println(string(msg.Data), "client")
}

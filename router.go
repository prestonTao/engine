package engine

import (
	// "mandela/peerNode/messageEngine/net"
	"sync"
)

// var router = new(RouterStore)

type MsgHandler func(c Controller, msg Packet)

var handlersMapping = make(map[uint32]MsgHandler)
var router_lock = new(sync.RWMutex)

func AddRouter(msgId uint32, handler MsgHandler) {
	router_lock.Lock()
	defer router_lock.Unlock()
	handlersMapping[msgId] = handler
}

func GetHandler(msgId uint32) MsgHandler {
	router_lock.Lock()
	defer router_lock.Unlock()
	handler := handlersMapping[msgId]
	return handler
}

package engine

import (
	"sync"
)

type MsgHandler func(c Controller, msg Packet)

type Router struct {
	handlersMapping map[uint32]MsgHandler
	lock            *sync.RWMutex
}

func (this *Router) AddRouter(msgId uint32, handler MsgHandler) {
	this.lock.Lock()
	this.handlersMapping[msgId] = handler
	this.lock.Unlock()
}

func (this *Router) GetHandler(msgId uint32) MsgHandler {
	this.lock.RLock()
	handler := this.handlersMapping[msgId]
	this.lock.RUnlock()
	return handler
}

func NewRouter() *Router {
	return &Router{
		handlersMapping: make(map[uint32]MsgHandler),
		lock:            new(sync.RWMutex),
	}
}

// var handlersMapping = make(map[uint32]MsgHandler)
// var router_lock = new(sync.RWMutex)

// func AddRouter(msgId uint32, handler MsgHandler) {
// 	router_lock.Lock()
// 	defer router_lock.Unlock()
// 	handlersMapping[msgId] = handler
// }

// func GetHandler(msgId uint32) MsgHandler {
// 	router_lock.RLock()
// 	defer router_lock.RUnlock()
// 	handler := handlersMapping[msgId]
// 	return handler
// }

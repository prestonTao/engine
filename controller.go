package engine

import (
	"sync"
)

type Controller interface {
	GetSession(name string) (Session, bool)      //通过accId得到客户端的连接Id
	GetNet() *Net                                //获得连接到本地的计算机连接
	SetAttribute(name string, value interface{}) //设置共享数据，实现业务模块之间通信
	GetAttribute(name string) interface{}        //得到共享数据，实现业务模块之间通信
}

type ControllerImpl struct {
	lock       *sync.RWMutex
	net        *Net
	attributes map[string]interface{}
}

//得到net模块，用于给用户发送消息
func (this *ControllerImpl) GetNet() *Net {
	return this.net
}

func (this *ControllerImpl) SetAttribute(name string, value interface{}) {
	this.lock.Lock()
	this.attributes[name] = value
	this.lock.Unlock()
}
func (this *ControllerImpl) GetAttribute(name string) interface{} {
	this.lock.RLock()
	itr := this.attributes[name]
	this.lock.RUnlock()
	return itr
}

//
func (this *ControllerImpl) GetSession(name string) (Session, bool) {
	this.lock.RLock()
	s, ok := this.net.GetSession(name)
	this.lock.RUnlock()
	return s, ok
}

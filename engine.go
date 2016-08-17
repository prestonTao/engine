package engine

import (
	"sync"
)

// var (
// 	Name       = "myserver"
// 	IP         = "127.0.0.1"
// 	PORT int32 = 9090
// )

type Engine struct {
	name     string
	status   int //服务器状态
	net      *Net
	auth     Auth
	onceRead *sync.Once
}

/*
	注册一个普通消息
*/
func (this *Engine) RegisterMsg(msgId uint32, handler MsgHandler) {
	this.net.router.AddRouter(msgId, handler)
}

func (this *Engine) Listen(ip string, port int32) error {
	return this.net.Listen(ip, port)
}

/*
	添加一个连接，给这个连接取一个名字，连接名字可以在自定义权限验证方法里面修改
	@powerful      是否是强连接
	@return  name  对方的名称
*/
func (this *Engine) AddClientConn(ip string, port int32, powerful bool) (name string, err error) {
	session, err := this.net.AddClientConn(ip, this.name, port, powerful)
	if err != nil {
		return "", err
	}
	name = session.GetName()
	return name, nil
}

//给一个session绑定另一个名称
func (this *Engine) LinkName(name string, session Session) {

}

//添加一个拦截器，所有消息到达业务方法之前都要经过拦截器处理
func (this *Engine) AddInterceptor(itpr Interceptor) {
	this.net.interceptor.addInterceptor(itpr)
}

//得到控制器
//func (this *Engine) GetController() Controller {
//	return this.controller
//}

//获得session
func (this *Engine) GetSession(name string) (Session, bool) {
	return this.net.GetSession(name)
}

//设置自定义权限验证
func (this *Engine) SetAuth(auth Auth) {
	if auth == nil {
		return
	}
	defaultAuth = auth
}

//func (this *Engine) SetInPacket(packet GetPacket) {
//	this.net.inPacket = packet
//}

//func (this *Engine) SetOutPacket(packet GetPacketBytes) {
//	this.net.outPacket = packet
//}

//设置关闭连接回调方法
func (this *Engine) SetCloseCallback(call CloseCallback) {
	this.net.closecallback = call
}

//@name   本服务器名称
func NewEngine(name string) *Engine {
	engine := new(Engine)
	engine.name = name
	//	engine.interceptor = NewInterceptor()
	engine.onceRead = new(sync.Once)
	engine.net = NewNet(name)
	return engine
}

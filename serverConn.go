package engine

import (
	// "fmt"
	"net"
	"sync"
)

//其他计算机对本机的连接
type ServerConn struct {
	sessionBase
	conn           net.Conn
	Ip             string
	Connected_time string
	CloseTime      string
	packet         Packet
	//	inPack         chan *Packet
	isClose    bool //该连接是否已经关闭
	net        *Net
	controller Controller
}

func (this *ServerConn) run() {
	this.controller = &ControllerImpl{
		lock:       new(sync.RWMutex),
		net:        this.net,
		attributes: make(map[string]interface{}),
	}
	this.packet.Session = this

	go this.recv()
}

//接收客户端消息协程
func (this *ServerConn) recv() {
	defer PrintPanicStack()
	//处理客户端主动断开连接的情况
	for !this.isClose {
		n, err := this.conn.Read(this.tempcache)
		if err != nil {
			this.Close()
			break
		}
		//TODO 判断超过16k的情况，断开客户端
		copy(this.cache, append(this.cache[:this.cacheindex], this.tempcache[:n]...))
		this.cacheindex = this.cacheindex + uint32(n)

		var ok bool
		var handler MsgHandler
		for {
			err, ok = RecvPackage(&this.cache, &this.cacheindex, &this.packet)
			if !ok {
				if err != nil {
					this.isClose = true
					Log.Warn("net error %s", err.Error())
					this.Close()
					return
				}
				break
			} else {
				Log.Debug("conn recv: %d, %s, %d", this.packet.MsgID, this.conn.RemoteAddr(), len(this.packet.Data))

				handler = this.net.router.GetHandler(this.packet.MsgID)
				if handler == nil {
					Log.Warn("该消息未注册，消息编号：%d", this.packet.MsgID)
				} else {
					//这里决定了消息是否异步处理
					this.handlerProcess(handler, &this.packet)
				}

				copy(this.cache, this.cache[this.packet.Size:this.cacheindex])
				this.cacheindex = this.cacheindex - this.packet.Size

			}
		}
	}

	//	this.net.CloseClient(this.GetName())
	this.Close()
	//最后一个包接收了之后关闭chan
	//如果有超时包需要等超时了才关闭，目前未做处理
	// close(this.outData)
	// fmt.Println("关闭连接")
}

func (this *ServerConn) handlerProcess(handler MsgHandler, msg *Packet) {
	//消息处理模块报错将不会引起宕机
	defer PrintPanicStack()
	//消息处理前先通过拦截器
	itps := this.net.interceptor.getInterceptors()
	itpsLen := len(itps)
	for i := 0; i < itpsLen; i++ {
		isIntercept := itps[i].In(this.controller, *msg)
		//
		if isIntercept {
			return
		}
	}
	handler(this.controller, *msg)
	//消息处理后也要通过拦截器
	for i := itpsLen; i > 0; i-- {
		itps[i-1].Out(this.controller, *msg)
	}
}

//给客户端发送数据
func (this *ServerConn) Send(msgID, opt, errcode uint32, cryKey []byte, data *[]byte) (err error) {
	defer PrintPanicStack()
	buff := MarshalPacket(msgID, opt, errcode, cryKey, data)
	//	index := 0
	//	for {
	//		if len(*buff) > 1024 {
	//			_, err = this.conn.Write((*buff)[index : index+1024])
	//			index = index + 1024
	//		} else {
	//			_, err = this.conn.Write((*buff)[index:])
	//			break
	//		}
	//	}
	_, err = this.conn.Write(*buff)
	Log.Debug("conn send: %d, %s, %d", msgID, this.conn.RemoteAddr(), len(*buff))
	return
}

//关闭这个连接
func (this *ServerConn) Close() {
	// fmt.Println("调用关闭连接方法")
	this.isClose = true
	err := this.conn.Close()
	if err != nil {
	}
	this.sessionStore.removeSession(this.GetName())
}

//获取远程ip地址和端口
func (this *ServerConn) GetRemoteHost() string {
	return this.conn.RemoteAddr().String()
}

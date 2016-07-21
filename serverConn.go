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
	inPack         chan *Packet
	isClose        bool //该连接是否已经关闭
	net            *Net
	controller     Controller
}

func (this *ServerConn) run() {
	this.controller = &ControllerImpl{
		lock: new(sync.RWMutex),
		net:  this.net,
		//		engine:              this.net,
		attributes: make(map[string]interface{}),
		//		msgGroup:            NewMsgGroupManager(),
		//		msgGroup.controller: c,
	}

	go this.recv()
	// go this.packetRouter()
	// go this.send()
	// go this.hold()
}

//接收客户端消息协程
func (this *ServerConn) recv() {
	//处理客户端主动断开连接的情况

	// n, err := conn.Read(header)
	// temp := make([]byte, 1024)
	for !this.isClose {
		n, err := this.conn.Read(this.tempcache)
		if err != nil {
			this.Close()
			break
		}
		//TODO 判断超过16k的情况，断开客户端
		copy(this.cache, append(this.tempcache[:this.cacheindex], this.tempcache[:n]...))
		// if this.cacheindex == 0 {
		// } else {
		// 	this.cache = append(this.cache, this.tempcache[:n]...)
		// }
		this.cacheindex = this.cacheindex + uint32(n)

		// fmt.Println("收到n个byte ", n, this.cache[:100], this.cacheindex)

		for {
			packet, err := RecvPackage(&this.cache, &this.cacheindex)
			if packet == nil {
				if err != nil {
					this.isClose = true
					Log.Warn("net error %s", err.Error())
					// panic(err)
				}
				break
			} else {
				// if packet.MsgID == 0 {
				// 	//hold 心跳包
				// 	continue
				// }
				packet.Session = this
				// this.inPack <- packet
				handler := GetHandler(packet.MsgID)
				if handler == nil {
					Log.Warn("该消息未注册，消息编号：%d", packet.MsgID)
				} else {
					// Log.Debug("conn recv: %d %s", packet.MsgID, string(packet.Data))
					// fmt.Println("recv   data", packet.Data)
					//这里决定了消息是否异步处理
					this.handlerProcess(handler, packet)
				}

				copy(this.cache, this.cache[packet.Size:this.cacheindex])
				this.cacheindex = this.cacheindex - packet.Size
				// fmt.Println("复制cache", t, this.cache[:100], this.cacheindex)

			}
			// fmt.Println("接收数据出错  ", err.Error())
			// Log.Debug("接收数据出错  %v", err)
		}
	}

	this.net.CloseClient(this.GetName())
	//最后一个包接收了之后关闭chan
	//如果有超时包需要等超时了才关闭，目前未做处理
	// close(this.outData)
	// fmt.Println("关闭连接")
}

/*
	路由接收过来的包
*/
// func (this *ServerConn) packetRouter() {
// 	for msg := range this.inPack {
// 		handler := GetHandler(msg.MsgID)
// 		if handler == nil {
// 			Log.Warn("该消息未注册，消息编号：%d", msg.MsgID)
// 			return
// 		}
// 		//这里决定了消息是否异步处理
// 		this.handlerProcess(handler, msg)
// 	}
// }

func (this *ServerConn) handlerProcess(handler MsgHandler, msg *Packet) {
	//消息处理模块报错将不会引起宕机
	//	defer func() {
	//		if err := recover(); err != nil {
	//			e, ok := err.(error)
	//			if ok {
	//				Log.Error("handler error msgId:%d \n %v", msg.MsgID, e)
	//				//				fmt.Println("网络库：", e.Error())
	//			}
	//		}
	//	}()
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

//发送给客户端消息协程
// func (this *ServerConn) send() {
// 	//处理客户端主动断开连接的情况
// 	//确保消息发送完后再关闭连接
// 	for msg := range this.outData {
// 		if _, err := this.conn.Write(*msg); err != nil {
// 			log.Println("发送数据出错", err)
// 			return
// 		}
// 	}
// }

// //心跳连接
// func (this *ServerConn) hold() {
// 	for !this.isClose {
// 		time.Sleep(time.Second * 5)
// 		bs := []byte("")
// 		this.Send(0, &bs)
// 	}
// }

//给客户端发送数据
func (this *ServerConn) Send(msgID, opt, errcode uint32, cryKey []byte, data *[]byte) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err, _ = e.(error)
		}
	}()
	buff := MarshalPacket(msgID, opt, errcode, cryKey, data)
	// this.outData <- buff
	_, err = this.conn.Write(*buff)
	Log.Debug("conn send: %d, %s, %d", msgID, this.conn.RemoteAddr(), len(*buff))
	return
}

//关闭这个连接
func (this *ServerConn) Close() {
	// fmt.Println("调用关闭连接方法")
	this.isClose = true
	this.Send(CloseConn, 0, 0, []byte{}, &zero_bytes)
}

//获取远程ip地址和端口
func (this *ServerConn) GetRemoteHost() string {
	return this.conn.RemoteAddr().String()
}

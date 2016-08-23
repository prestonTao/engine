package engine

import (
	"net"
	"strconv"
	"sync"
	"time"
)

//本机向其他服务器的连接
type Client struct {
	sessionBase
	serverName string
	ip         string
	port       int32
	conn       net.Conn
	inPack     chan *Packet //接收队列
	packet     Packet       //
	isClose    bool         //该连接是否被关闭
	isPowerful bool         //是否是强连接，强连接有短线重连功能
	net        *Net
	controller Controller
}

func (this *Client) Connect(ip string, port int32) (remoteName string, err error) {

	this.ip = ip
	this.port = port

	this.conn, err = net.Dial("tcp", ip+":"+strconv.Itoa(int(port)))
	if err != nil {
		return
	}

	//权限验证
	remoteName, err = defaultAuth.SendKey(this.conn, this, this.serverName)
	if err != nil {
		return
	}

	Log.Debug("Connecting to %s:%s", ip, strconv.Itoa(int(port)))

	this.controller = &ControllerImpl{
		lock:       new(sync.RWMutex),
		net:        this.net,
		attributes: make(map[string]interface{}),
	}
	this.packet.Session = this
	go this.recv()
	return
}
func (this *Client) reConnect() {
	for {
		//十秒钟后重新连接
		time.Sleep(time.Second * 10)
		var err error
		this.conn, err = net.Dial("tcp", this.ip+":"+strconv.Itoa(int(this.port)))
		if err != nil {
			continue
		}

		Log.Debug("Connecting to %s:%s", this.ip, strconv.Itoa(int(this.port)))

		go this.recv()
		return
	}
}

func (this *Client) recv() {
	defer PrintPanicStack()
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

	this.net.CloseClient(this.GetName())
	if this.isPowerful {
		go this.reConnect()
	}

}

func (this *Client) handlerProcess(handler MsgHandler, msg *Packet) {
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

//发送序列化后的数据
func (this *Client) Send(msgID, opt, errcode uint32, cryKey []byte, data *[]byte) (err error) {
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

//客户端关闭时,退出recv,send
func (this *Client) Close() {
	this.isClose = true
	err := this.conn.Close()
	if err != nil {
	}
	this.sessionStore.removeSession(this.GetName())
}

//获取远程ip地址和端口
func (this *Client) GetRemoteHost() string {
	return this.conn.RemoteAddr().String()
}

func NewClient(name, ip string, port int32) *Client {
	client := new(Client)
	client.name = name
	client.inPack = make(chan *Packet, 1000)
	// client.outData = make(chan *[]byte, 1000)
	client.Connect(ip, port)
	return client
}

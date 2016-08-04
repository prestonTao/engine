package engine

import (
	"net"
	"strconv"
	"sync"
	"time"
)

func init() {
	// GlobalInit("console", "", "debug", 1)
	// utils.GlobalInit("file", `{"filename":"/var/log/gd/gd.log"}`, "", 1000)
	// utils.Log.Debug("session handle receive, %d, %v", msg.Code(), msg.Content())
	//	Log.Debug("test debug")
	//	Log.Warn("test warn")
	//	Log.Error("test error")
}

type Net struct {
	Name          string //本机名称
	interceptor   *InterceptorProvider
	sessionStore  *sessionStore
	closecallback CloseCallback
	ipPort        string
	lis           *net.TCPListener
	router        *Router
	inPacket      GetPacket
	outPacket     GetPacketBytes
	isSuspend     bool
}

func (this *Net) Listen(ip string, port int32) error {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", ip+":"+strconv.Itoa(int(port)))
	if err != nil {
		Log.Error("%v", err)
		return err
	}

	this.lis, err = net.ListenTCP("tcp4", tcpAddr)
	if err != nil {
		Log.Error("%v", err)
		return err
	}
	Log.Debug("监听一个地址：%s", ip+":"+strconv.Itoa(int(port)))
	go this.listener(this.lis)
	return nil
}

func (this *Net) listener(listener *net.TCPListener) {
	this.lis = listener
	this.ipPort = listener.Addr().String()
	for !this.isSuspend {
		conn, err := this.lis.Accept()
		if err != nil {
			continue
		}
		if this.isSuspend {
			conn.Close()
			continue
		}
		go this.newConnect(conn)
	}
}

//创建一个新的连接
func (this *Net) newConnect(conn net.Conn) {
	defer PrintPanicStack()
	remoteName, err := defaultAuth.RecvKey(conn, this.Name)
	if err != nil {
		return
	}

	sessionBase := sessionBase{
		cache:      make([]byte, 16*1024*1024, 16*1024*1024),
		cacheindex: 0,
		tempcache:  make([]byte, 1024, 1024),
		lock:       new(sync.RWMutex),
	}

	serverConn := &ServerConn{
		sessionBase:    sessionBase,
		conn:           conn,
		Ip:             conn.RemoteAddr().String(),
		Connected_time: time.Now().String(),
		inPack:         make(chan *Packet, 500),
		net:            this,
	}
	serverConn.sessionStore = this.sessionStore
	serverConn.name = remoteName
	serverConn.attrbutes = make(map[string]interface{}, 10)
	serverConn.run()
	this.sessionStore.addSession(remoteName, serverConn)

	// fmt.Println(time.Now().String(), "建立连接", conn.RemoteAddr().String())
	// Log.Debug("%s 建立连接：%s", time.Now().String(), conn.RemoteAddr().String())

}

//关闭连接
func (this *Net) CloseClient(name string) bool {
	session, ok := this.sessionStore.getSession(name)
	if ok {
		if this.closecallback != nil {
			this.closecallback(name)
		}
		this.sessionStore.removeSession(name)
		session.Close()
		return true
	}
	return false
}

/*
	连接一个服务器
	@serverName   给客户端发送的自己的名字
	@powerful     是否是强连接，是强连接断开后自动重连
*/
func (this *Net) AddClientConn(ip, serverName string, port int32, powerful bool) (Session, error) {
	sessionBase := sessionBase{
		cache:      make([]byte, 1024, 16*1024*1024),
		cacheindex: 0,
		tempcache:  make([]byte, 1024, 1024),
		lock:       new(sync.RWMutex),
	}
	clientConn := &Client{
		sessionBase: sessionBase,
		serverName:  serverName,
		inPack:      make(chan *Packet, 5000),
		net:         this,
		isPowerful:  powerful,
	}
	clientConn.sessionStore = this.sessionStore
	clientConn.attrbutes = make(map[string]interface{})
	remoteName, err := clientConn.Connect(ip, port)
	if err == nil {
		clientConn.name = remoteName
		this.sessionStore.addSession(remoteName, clientConn)
		return clientConn, nil
	}
	return nil, err
}

func (this *Net) GetSession(name string) (Session, bool) {
	return this.sessionStore.getSession(name)
}

func (this *Net) GetAllSession() []Session {
	return this.sessionStore.getAllSession()
}

//发送数据
func (this *Net) Send(name string, msgID, opt, errcode uint32, cryKey []byte, data []byte) bool {
	session, ok := this.sessionStore.getSession(name)
	if ok {
		session.Send(msgID, opt, errcode, cryKey, &data)
		return true
	} else {
		return false
	}
}

//@name   本服务器名称
func NewNet(name string) *Net {
	net := new(Net)
	net.Name = name
	net.interceptor = NewInterceptor()
	net.sessionStore = NewSessionStore()
	net.inPacket = RecvPackage
	net.outPacket = MarshalPacket
	net.router = NewRouter()
	return net
}

/*
	暂停服务器
*/
func (this *Net) Suspend(names ...string) {
	Log.Debug("暂停服务器")
	// this.lis.Close()
	this.isSuspend = true
	for _, one := range this.sessionStore.getAllSessionName() {
		done := false
		for _, nameOne := range names {
			if nameOne == one {
				done = true
				break
			}
		}
		if done {
			continue
		}
		if session, ok := this.GetSession(one); ok {
			session.Close()
		}
	}
}

/*
	恢复服务器
*/
func (this *Net) Recovery() {
	Log.Debug("恢复服务器")
	this.isSuspend = false
	go this.listener(this.lis)
}

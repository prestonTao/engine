package engine

import (
	// "errors"
	// "fmt"
	"sync"
)

type sessionBase struct {
	sessionStore *sessionStore
	name         string
	attrbutes    map[string]interface{}
	cache        []byte
	cacheindex   uint32
	tempcache    []byte
	lock         *sync.RWMutex
}

func (this *sessionBase) Set(name string, value interface{}) {
	this.lock.Lock()
	this.attrbutes[name] = value
	this.lock.Unlock()
}
func (this *sessionBase) Get(name string) interface{} {
	this.lock.RLock()
	itr := this.attrbutes[name]
	this.lock.RUnlock()
	return itr
}
func (this *sessionBase) GetName() string {
	return this.name
}
func (this *sessionBase) SetName(name string) {
	this.sessionStore.renameSession(this.name, name)
	this.name = name
}
func (this *sessionBase) Send(msgID, opt, errcode uint32, cryKey []byte, data *[]byte) (err error) {
	return
}
func (this *sessionBase) Close() {}
func (this *sessionBase) GetRemoteHost() string {
	return "127.0.0.1:0"
}

type Session interface {
	Send(msgID, opt, errcode uint32, cryKey []byte, data *[]byte) error
	Close()
	Set(name string, value interface{})
	Get(name string) interface{}
	GetName() string
	SetName(name string)
	GetRemoteHost() string
}

type sessionStore struct {
	lock *sync.RWMutex
	// store     map[int64]Session
	nameStore map[string]Session
}

func (this *sessionStore) addSession(name string, session Session) {
	this.lock.Lock()
	this.nameStore[session.GetName()] = session
	this.lock.Unlock()
}

func (this *sessionStore) getSession(name string) (Session, bool) {
	this.lock.RLock()
	s, ok := this.nameStore[name]
	this.lock.RUnlock()
	return s, ok
}

func (this *sessionStore) removeSession(name string) {
	this.lock.Lock()
	delete(this.nameStore, name)
	this.lock.Unlock()
}

func (this *sessionStore) renameSession(oldName, newName string) {
	s, ok := this.nameStore[oldName]
	if !ok {
		return
	}
	this.lock.Lock()
	delete(this.nameStore, oldName)
	this.lock.Unlock()
	this.nameStore[newName] = s
}

func (this *sessionStore) getAllSession() []Session {
	this.lock.RLock()
	ss := make([]Session, 0)
	for _, s := range this.nameStore {
		ss = append(ss, s)
	}
	this.lock.RUnlock()
	return ss
}

func (this *sessionStore) getAllSessionName() []string {
	names := make([]string, 0)
	this.lock.RLock()
	for key, _ := range this.nameStore {
		names = append(names, key)
	}
	this.lock.RUnlock()
	return names
}

func NewSessionStore() *sessionStore {
	sessionStore := new(sessionStore)
	sessionStore.lock = new(sync.RWMutex)
	sessionStore.nameStore = make(map[string]Session)
	return sessionStore
}

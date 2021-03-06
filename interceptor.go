package engine

import (
	"sync"
)

// type interceptor struct{}

// func (this *interceptor) in() {

// }

type Interceptor interface {
	In(c Controller, msg Packet) bool
	Out(c Controller, msg Packet)
}

type InterceptorProvider struct {
	lock         *sync.RWMutex
	interceptors []Interceptor
}

func (this *InterceptorProvider) addInterceptor(itpr Interceptor) {
	this.lock.Lock()
	this.interceptors = append(this.interceptors, itpr)
	this.lock.Unlock()
}
func (this *InterceptorProvider) getInterceptors() []Interceptor {
	return this.interceptors

}

func NewInterceptor() *InterceptorProvider {
	interceptor := new(InterceptorProvider)
	interceptor.lock = new(sync.RWMutex)
	interceptor.interceptors = make([]Interceptor, 0)
	return interceptor
}

// var interceptors *InterceptorProvider

// func init() {
// 	interceptors = new(InterceptorProvider)
// 	interceptors.lock = new(sync.RWMutex)
// 	interceptors.interceptors = make([]Interceptor, 0)
// }

// var chanS =

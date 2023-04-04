package hamble

import (
	"fmt"
	"github.com/dawnzzz/hamble-tcp-server/iface"
	"sync"
)

type Router struct {
	apis map[uint32]iface.IHandler
	mu   sync.Mutex
}

func newRouter() iface.IRouter {
	return &Router{
		apis: make(map[uint32]iface.IHandler),
	}
}

func (r *Router) AddRouter(id uint32, handler iface.IHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exist := r.apis[id]; exist {
		// 不能重复定义
		panic(fmt.Sprintf("%v handler register duplicate", id))
	}

	r.apis[id] = handler
}

func (r *Router) GetHandler(id uint32) iface.IHandler {
	r.mu.Lock()
	defer r.mu.Unlock()

	if handler, exist := r.apis[id]; exist {
		return handler
	}

	return &BaseHandler{}
}

func (r *Router) DoHandler(request iface.IRequest) {
	handler := r.GetHandler(request.GetMsgID()) // 根据MsgID获取handler

	handler.PreHandle(request)
	handler.Handle(request)
	handler.PostHandle(request)
}

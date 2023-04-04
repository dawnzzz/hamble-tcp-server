package iface

type IRouter interface {
	AddRouter(id uint32, handler IHandler) // 注册路由
	GetHandler(id uint32) IHandler         // 根据id获取handler
	DoHandler(request IRequest)
}

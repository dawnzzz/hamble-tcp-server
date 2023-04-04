package iface

type IHandler interface {
	PreHandle(request IRequest)  // 主业务服务之前的处理
	Handle(request IRequest)     // 主业务服务
	PostHandle(request IRequest) // 主业务服务之后的处理
}

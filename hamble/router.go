package hamble

import (
	"fmt"
	"github.com/dawnzzz/hamble-tcp-server/conf"
	"github.com/dawnzzz/hamble-tcp-server/iface"
	"github.com/dawnzzz/hamble-tcp-server/logger"
	"sync"
)

type Router struct {
	apis map[uint32]iface.IHandler
	mu   sync.Mutex

	workerPoolSize int                   // worker 的数量
	taskQueues     []chan iface.IRequest // Worker 负责取任务的消息队列
}

func newRouter() iface.IRouter {
	return &Router{
		apis: make(map[uint32]iface.IHandler),

		workerPoolSize: conf.GlobalProfile.WorkerPoolSize,
		taskQueues:     make([]chan iface.IRequest, conf.GlobalProfile.WorkerPoolSize),
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

func (r *Router) startOneWorker(workerID int, taskQueue chan iface.IRequest) {
	logger.Infof("Worker ID = %v is started", workerID)

	//不断等待队列中的消息
	for {
		select {
		//有消息则取出队列的Request，并执行绑定的业务方法
		case request := <-taskQueue:
			r.DoHandler(request)
		}
	}
}

func (r *Router) StartWorkerPool() {
	for i := 0; i < r.workerPoolSize; i++ {
		//给当前worker对应的任务队列开辟空间
		r.taskQueues[i] = make(chan iface.IRequest, conf.GlobalProfile.MaxWorkerTaskLen)

		// 开启worker
		go r.startOneWorker(i, r.taskQueues[i])
	}
}

func (r *Router) SendMsgToTaskQueue(request iface.IRequest) {
	//根据ConnID来分配当前的连接应该由哪个worker负责处理

	workerID := int(request.GetMsgID()) % r.workerPoolSize
	logger.Infof("Add request msgID=%v to workerID=%v", request.GetMsgID(), workerID)
	//将请求消息发送给任务队列
	r.taskQueues[workerID] <- request
}

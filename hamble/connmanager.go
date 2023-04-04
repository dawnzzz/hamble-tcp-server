package hamble

import (
	"github.com/dawnzzz/hamble-tcp-server/iface"
	"github.com/dawnzzz/hamble-tcp-server/logger"
	"sync"
	"sync/atomic"
)

type ConnManager struct {
	connections map[iface.IConnection]struct{}

	mu         sync.Mutex
	isClearing atomic.Bool
}

func NewConnManager() iface.IConnManager {
	return &ConnManager{
		connections: make(map[iface.IConnection]struct{}),
	}
}

func (cm *ConnManager) Add(connection iface.IConnection) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	//将conn连接添加到ConnManager中
	cm.connections[connection] = struct{}{}

	logger.Infof("connection add to ConnManager successfully: conn num = %v", cm.Len())
}

func (cm *ConnManager) Remove(connection iface.IConnection) {
	if cm.isClearing.Load() {
		// 正在进行清除操作，直接返回，由Clear方法删除连接
		return
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	delete(cm.connections, connection)

	logger.Infof("connection remove from ConnManager successfully: conn num = %v", cm.Len())
}

func (cm *ConnManager) Len() int {
	return len(cm.connections)
}

func (cm *ConnManager) Clear() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.isClearing.Store(true)

	for connection := range cm.connections {
		// 停止
		connection.Stop()
		// 删除
		delete(cm.connections, connection)
	}

	logger.Infof("Conn Manager cleared: conn num = %v", cm.Len())
}

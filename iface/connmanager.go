package iface

// IConnManager 链接管理的抽象表示
type IConnManager interface {
	Add(connection IConnection)    // 添加链接
	Remove(connection IConnection) // 删除链接
	Len() int                      // 获取链接个数
	Clear()                        // 清除所有的链接
}

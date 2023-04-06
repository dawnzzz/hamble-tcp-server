package iface

// ICSBase ISserver和IClient的祖先，这两个接口都继承于此
type ICSBase interface {
	RegisterHandler(id uint32, handler IHandler) // 注册Handler
	GetRouter() IRouter                          // 获取Router
	GetConnManager() IConnManager                // 获取ConnManager
	SetOnConnStart(func(conn IConnection))       // 设置连接创建时的Hook函数
	SetOnConnStop(func(conn IConnection))        // 设置连接结束时的Hook函数
	CallOnConnStart(conn IConnection)            // 调用连接创建时的Hook函数
	CallOnConnStop(conn IConnection)             // 调用连接结束时的Hook函数
	GetHeartBeatChecker() IHeartBeatChecker      // 获取心跳检测器
	GetDataPack() IDataPack                      // 获取封包/解包方式
	SetDataPack(dataPack IDataPack)              // 设置封包/解包方式
}

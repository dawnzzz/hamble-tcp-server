package iface

type IMessage interface {
	GetMsgID() uint32   // 获取消息ID
	GetData() []byte    // 获取数据
	GetDataLen() uint32 // 获取数据头长度

	SetMsgID(id uint32)       // 设置消息ID
	SetData(data []byte)      // 设置数据包数据
	SetDataLen(length uint32) // 设置数据长度
}

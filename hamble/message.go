package hamble

import "github.com/dawnzzz/hamble-tcp-server/iface"

type Message struct {
	msgID  uint32
	data   []byte
	length uint32
}

func NewMessage(msgID uint32, data []byte) iface.IMessage {
	return &Message{
		msgID:  msgID,
		data:   data,
		length: uint32(len(data)),
	}
}

func (m *Message) GetMsgID() uint32 {
	return m.msgID
}

func (m *Message) GetData() []byte {
	return m.data
}

func (m *Message) GetDataLen() uint32 {
	return m.length
}

func (m *Message) SetMsgID(msgID uint32) {
	m.msgID = msgID
}

func (m *Message) SetData(data []byte) {
	m.data = data
}

func (m *Message) SetDataLen(length uint32) {
	m.length = length
}

package hamble

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/dawnzzz/hamble-tcp-server/conf"
	"github.com/dawnzzz/hamble-tcp-server/iface"
)

const headLen = uint32(8) // 数据长度4字节（uint32）+MsgID占4字节（uint32）

type DataPack struct {
}

func NewDataPack() iface.IDataPack {
	return &DataPack{}
}

func (dp *DataPack) GetHeadLen() uint32 {
	return headLen
}

// Pack 封包方法
func (dp *DataPack) Pack(msg iface.IMessage) ([]byte, error) {
	buffer := bytes.NewBuffer([]byte{})

	// 写入数据长度
	err := binary.Write(buffer, binary.BigEndian, msg.GetDataLen())
	if err != nil {
		return nil, err
	}

	// 写入MsgID
	err = binary.Write(buffer, binary.BigEndian, msg.GetMsgID())
	if err != nil {
		return nil, err
	}

	// 写入数据
	err = binary.Write(buffer, binary.BigEndian, msg.GetData())
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// Unpack 解包方法
func (dp *DataPack) Unpack(data []byte) (iface.IMessage, error) {
	reader := bytes.NewReader(data)

	msg := &Message{}
	// 读取数据长度
	err := binary.Read(reader, binary.BigEndian, &msg.length)
	if err != nil {
		return nil, err
	}

	// 读取MsgID
	err = binary.Read(reader, binary.BigEndian, &msg.msgID)
	if err != nil {
		return nil, err
	}

	//判断dataLen的长度是否超出允许的最大包长度
	if conf.GlobalProfile.MaxPacketSize > 0 && msg.length > conf.GlobalProfile.MaxPacketSize {
		return nil, errors.New("packet size is too big")
	}

	return msg, nil
}

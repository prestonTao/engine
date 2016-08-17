package engine

import (
	"bytes"
	"crypto/rc4"
	"encoding/binary"
	"errors"
	// "fmt"
)

// var defaultGetPacket GetPacket = RecvPackage

type GetPacket func(cache *[]byte, index *uint32) (packet *Packet, e error)
type GetPacketBytes func(msgID, opt, errcode uint32, cryKey []byte, data *[]byte) *[]byte

type Packet struct {
	Opt       uint32
	MsgID     uint32
	Size      uint32
	Errorcode uint32
	Data      []byte
	Crypt_key []byte
	Session   Session
}

/*
	系统默认的消息接收并转化为Packet的方法
	一个packet包括包头和包体，保证在接收到包头后两秒钟内接收到包体，否则线程会一直阻塞
	因此，引入了超时机制
*/
func RecvPackage(cache *[]byte, index *uint32, packet *Packet) (error, bool) {
	// fmt.Println("packet   11111", *index, (*cache))

	if *index < 24 {
		return nil, false
	}
	//	packet = new(Packet)
	packet.Opt = binary.LittleEndian.Uint32((*cache)[:4])
	packet.Size = binary.LittleEndian.Uint32((*cache)[4:8])
	if packet.Size < 24 {
		// fmt.Println(*cache)
		return errors.New("包头错误"), false
	}
	packet.MsgID = binary.LittleEndian.Uint32((*cache)[8:12])
	packet.Errorcode = binary.LittleEndian.Uint32((*cache)[12:16])
	packet.Crypt_key = (*cache)[16:24]

	// bodyBytes := make([]byte, packet.Size-24)

	// timeout := NewTimeOut(func() {

	// n, err = io.ReadFull(conn, bodyBytes)
	// fmt.Println(*index, packet.Size)
	if *index < packet.Size {
		return nil, false
	}
	// fmt.Println("packet   3333333", packet.Size)
	// fmt.Println((*cache)[24 : packet.Size+2])
	packet.Data = (*cache)[24:packet.Size]
	// fmt.Println("packet  data 1", packet.Data)

	if (packet.Opt & 0x00800000) != 0 {
		packet.Crypt_key = cry(packet.Crypt_key)
		// key := []byte{1, 2, 3, 4, 5, 6, 7}
		c, err := rc4.NewCipher(packet.Crypt_key)
		if err != nil {
			// log.Println("", err)
			// Log.Debug("rc4 error: %v", err)

			return errors.New("rc4 error " + err.Error()), false
		}
		// fmt.Println("packet  加解密")
		c.XORKeyStream(packet.Data, packet.Data)
	}

	// fmt.Println("packet  data 2", packet.Data)
	return nil, true
}

func MarshalPacket(msgID, opt, errcode uint32, cryKey []byte, data *[]byte) *[]byte {
	newCryKey := cryKey
	if data == nil {
		buf := bytes.NewBuffer([]byte{})
		binary.Write(buf, binary.LittleEndian, opt)
		binary.Write(buf, binary.LittleEndian, uint32(24))
		binary.Write(buf, binary.LittleEndian, msgID)
		binary.Write(buf, binary.LittleEndian, errcode)
		buf.Write(newCryKey)
		bs := buf.Bytes()
		return &bs
	}
	bodyBytes := *data
	if (opt & 0x00800000) != 0 {
		newCryKey = cry(newCryKey)
		c, err := rc4.NewCipher(newCryKey)
		if err != nil {
			// log.Println("rc4 加密 error:", err)
			Log.Debug("rc4 加密 error: %v", err)
		}
		c.XORKeyStream(bodyBytes, bodyBytes)
	}
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, binary.LittleEndian, opt)
	binary.Write(buf, binary.LittleEndian, uint32(24+len(*data)))
	binary.Write(buf, binary.LittleEndian, msgID)
	binary.Write(buf, binary.LittleEndian, errcode)
	buf.Write(newCryKey)
	buf.Write(bodyBytes)
	bs := buf.Bytes()
	return &bs
}

func cry(in []byte) []byte {
	i := 0
	tmpBuf := make([]byte, 128)
	for i < len(in) {
		if i+1 < len(in) {
			tmpBuf[i] = in[i+1]
			tmpBuf[i+1] = in[i]
		} else {
			tmpBuf[i] = in[i]
		}
		i += 2
	}
	out := make([]byte, len(in))
	for i := 0; i < len(in); i++ {
		out[i] = tmpBuf[i] & 0x01
		out[i] = tmpBuf[i] & 0x0f
		out[i] <<= 4
		out[i] |= ((tmpBuf[i] & 0xf0) >> 4)
	}
	return out
}

package convert

import (
	"bytes"
	"fmt"
	"net"
	"time"

	"google.golang.org/protobuf/proto"
)

type MessagePag struct {
	c_type	uint8
	s_type	uint8
	msgID	uint16
}

func pagFilter(rep_pag_buffer *bytes.Buffer, target_id int) ([]byte, error) {
	// -1的意思是不在乎targetid，无需验证
	if target_id == -1 {
		return nil, nil
	}
	for {
		// if len()
		head := make([]byte, 4)
		rep_pag_buffer.Read(head)
		head = SliceRever(head)
		pag_len := BytesToUint32(head)
		pag_data := make([]byte, pag_len)
		rep_pag_buffer.Read(pag_data)
		pb_data, err := func() ([]byte, error) {
			// 做一些事情处理你的数据
			var pd []byte
			var e error
			return pb, e
		}()
		if msgID == target_id {
			pb_data_full := pag_data[5:]
			pb_compress_state := pb_data_full[:4]
			pb_compress_state = SliceRever(pb_compress_state)
			pb_data := pb_data_full[4:]
			pcs := BytesToUint32(pb_compress_state)
			// 正在使用新压缩方法
			if pcs != 0 {
				pb_data, err := UnzipLz4c(int(pcs), pb_data)
				return pb_data, err
			}
			return pb_data, nil
		}
	}
}

func PagCat(send_pag []byte, target_id int, conn net.Conn, rsp_pag_buffer bytes.Buffer) ([]byte, error) {
	var pb_data []byte
	
	// 发包
	_, err := conn.Write(send_pag)
	if err != nil {return pb_data, err}
	time.Sleep(800 * time.Millisecond)

	// 收包
	readTimeout := 2 * time.Second
	for {
		err = conn.SetReadDeadline(time.Now().Add(readTimeout))
		rsp_data := [10240]byte{}
		n, err := conn.Read(rsp_data[:])
		if err != nil {break}
		rsp_pag_buffer.Write(rsp_data[:n])
	}
	// 解包拿pb
	if rsp_pag_buffer.Len() == 0 {
		fmt.Println("收包为空")
	}
	pb_data, err = pagFilter(&rsp_pag_buffer, target_id)
	if err != nil {return pb_data, err}
	// if len(pb_data) == 0 {return pb_data, NewmsgError(pb_data)}

	return pb_data, nil
}

func PagData(pm proto.Message, fn string) ([]byte) {
	pb_data, _ := proto.Marshal(pm)
	mp := MessagePag{uint8(MapLit[fn][k1]), uint8(MapLit[fn][k2]), uint16(MapLit[fn][k3])}
	send_pag := mp.GetSendPag(pb_data)
	return send_pag
}

// 根据协议格式生成发送包的打包函数，需自行定义
func (mp *MessagePag)GetSendPag(pb_data []byte) []byte {
	buf := bytes.NewBuffer([]byte{})
	mI := SliceRever(Uint16ToBytes(mp.msgID))
	buf.Write(mI)
	buf.Write(pb_data)
	head := uint32(len(buf.Bytes()))
	sp :=SliceRever(Uint32ToBytes(head))
	sp = append(sp, buf.Bytes()...)
	buf.Reset()
	return sp
}
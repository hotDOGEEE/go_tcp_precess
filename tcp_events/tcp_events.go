package tcp_events

import (
	"bytes"
	"draft/ONE_MSG"
	"draft/convert"
	"net"
	"time"

	"google.golang.org/protobuf/proto"
)

type GateClient struct {
	conn           net.Conn
	rsp_pag_buffer bytes.Buffer
	// 这个主要用于接收必要的回包，并且往往含有基础数据，放在结构体中是一种较好的设计
	Login_rst      ONE_MSG.Login_Gate
}

type BaseClient struct {
	conn 			net.Conn
	rsp_pag_buffer 	bytes.Buffer

	Login_rst		ONE_MSG.Login_Base
}

func NewGateClient(conn net.Conn, rsp_pag_buffer bytes.Buffer, serial uint8) GateClient {
	var gc = GateClient{
		conn: conn,
		rsp_pag_buffer: rsp_pag_buffer,
	}
	return gc
}

func NewBaseClient(conn net.Conn, rsp_pag_buffer bytes.Buffer, serial uint8) BaseClient {
	var gc = BaseClient{
		conn: conn,
		rsp_pag_buffer: rsp_pag_buffer,
	}
	return gc
}

func (gate_c *GateClient) GateLoginEvent() error {
	// 第一个包，handshake
	handshake := func() error {
		send_pag := func() []byte {
			cgh := ONE_MSG.Handshake{
				ClientTime: uint64(time.Now().Unix()),
				ServerTime: uint64(0),
			}
			return convert.PagData(&cgh, string(cgh.ProtoReflect().Descriptor().FullName()))
		}()
		pb_data, err := convert.PagCat(send_pag, 1, gate_c.conn, gate_c.rsp_pag_buffer)
		if err != nil {
			return err
		}
		handshake := &ONE_MSG.HandshakeResp{}
		err = proto.Unmarshal(pb_data, handshake)
		if err != nil {
			return err
		}
		return nil
	}()
	if handshake != nil {
		return handshake
	}
	
	// loginEvent
	login := func() error {
		send_pag := func() []byte {
			clc := ONE_MSG.Login{
				Id:    uint64(525),
			}
			return convert.PagData(&clc, string(clc.ProtoReflect().Descriptor().FullName()))
		}()
		pb_data, err := convert.PagCat(send_pag, 1, gate_c.conn, gate_c.rsp_pag_buffer)
		
		if err != nil {
			return err
		}
		gate_c.Login_rst = ONE_MSG.LoginResp{}
		err = proto.Unmarshal(pb_data, &gate_c.Login_rst)
		if err != nil {
			return err
		}
		return nil
	}()
	if login != nil {
		return login
	}
	return nil
}

func (base_c *BaseClient) BaseLoginEvent() error {
	// handshake
	handshake := func() error {
		send_pag := func() []byte {
			cgh := ONE_MSG.Handshake{
				ClientTime: uint64(time.Now().Unix()),
				ServerTime: uint64(0),
			}
			return convert.PagData(&cgh, string(cgh.ProtoReflect().Descriptor().FullName()))
		}()
		pb_data, err := convert.PagCat(send_pag, 2, base_c.conn, base_c.rsp_pag_buffer)
		if err != nil {
			return err
		}
		handshake := &ONE_MSG.HandshakeResp{}
		err = proto.Unmarshal(pb_data, handshake)
		if err != nil {return err}
		return nil
	}()
	if handshake != nil {
		
		return handshake
	}
	
	// 补一个心跳
	heartbeat := base_c.heartbeat()
	if heartbeat != nil {return heartbeat}

	// clt_g_login
	login := func() error {
		send_pag := func() []byte {
			cgl := ONE_MSG.Login{
			}
			return convert.PagData(&cgl, string(cgl.ProtoReflect().Descriptor().FullName()))
		}()
		pb_data, err := convert.PagCat(send_pag, 3, base_c.conn, base_c.rsp_pag_buffer)
		if err != nil {return err}
		return nil
	}()
	if login != nil {
		return login
	}
	return nil
}

func (base_c *BaseClient) Chat(msg string) error {
	send_chatcontent := func() error {
		send_pag := func() []byte {
			cgs := ONE_MSG.ChatContent{
				Info: msg,
			}
			return convert.PagData(&cgs, string(cgs.ProtoReflect().Descriptor().FullName()))
		}()
		pb_data, err := convert.PagCat(send_pag, 4,
		base_c.conn, base_c.rsp_pag_buffer)
		if err != nil {return err}
		cco := &ONE_MSG.ChatResp{}
		err = proto.Unmarshal(pb_data, cco)
		if err != nil {return err}
		return nil
	}()
	if send_chatcontent != nil {return send_chatcontent}
	return nil
}

func (base_c *BaseClient) heartbeat() error {
	send_pag := func() []byte {
		cgh := ONE_MSG.Heartbeat{
			ClientTime: uint64(time.Now().Unix()),
			ServerTime: uint64(0),
		}
		return convert.PagData(&cgh, string(cgh.ProtoReflect().Descriptor().FullName()))
	}()
	pb_data, err := convert.PagCat(send_pag, 1, base_c.conn, base_c.rsp_pag_buffer)
	if err != nil {
		return err
	}
	heartbeat := &ONE_MSG.HeartbeatResp{}
	err = proto.Unmarshal(pb_data, heartbeat)

	if err != nil {
		return err
	}
	return nil
}

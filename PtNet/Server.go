package PtNet

/**
 * Title:TCP服务端接口
 * User: iuoon
 * Date: 2016-9-23
 * Version: 1.0
 */
import (
	"github.com/golang/protobuf/proto"
	"net"
	"sync"
)

import (
	"github.com/iuoon/PuetxGo/PtMsg"
	"github.com/iuoon/PuetxGo/PtUtil"
	"github.com/iuoon/PuetxGo/PtStatic"
)

//NewConnCallback 新连接回调
type NewConnCallback func(*GxTCPConn)

//DisConnCallback 连接断开回调
type DisConnCallback func(*GxTCPConn)

//clientMessageCallback 已经注册的命令回调
type clientMessageCallback func(*GxTCPConn, *PtMsg.GxMessage) error

//RawMessageCallback 没有注册的消息的回调
type RawMessageCallback func(*GxTCPConn, *PtMsg.GxMessage) error

//GxTCPServer tcp服务器
type GxTCPServer struct {
	mutex     *sync.Mutex
	addrConns map[string]*GxTCPConn //key为连接地址的map
	IDConns   map[uint32]*GxTCPConn //key为连接ID的map
	ID        uint32                //用来给客户端连接分配ID
	Mask      uint32                //id生成掩码

	Nc         NewConnCallback
	Dc         DisConnCallback
	Rm         RawMessageCallback
	clientCmds map[uint16]clientMessageCallback
}

//NewGxTCPServer 生成一个新的GxTCPServer
func NewGxTCPServer(mask uint32, nc NewConnCallback, dc DisConnCallback, rm RawMessageCallback, messageCtrl bool) *GxTCPServer {
	server := new(GxTCPServer)
	server.addrConns = make(map[string]*GxTCPConn)
	server.IDConns = make(map[uint32]*GxTCPConn)
	server.ID = 0
	server.mutex = new(sync.Mutex)
	server.Mask = mask

	server.Nc = nc
	server.Dc = dc
	server.Rm = rm
	server.clientCmds = make(map[uint16]clientMessageCallback)

	//注册心跳回调
	server.RegisterClientCmd(PtStatic.CmdHeartbeat, HeartbeatCallback)
	return server
}

//RegisterClientCmd 注册需要直接处理的消息
func (server *GxTCPServer) RegisterClientCmd(cmd uint16, cb clientMessageCallback) {
	_, ok := server.clientCmds[cmd]
	if ok {
		return
	}
	server.clientCmds[cmd] = cb
}

//Start 服务端启动函数
func (server *GxTCPServer) Start(port string) error {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		PtUtil.Debug("lister %s fail", port)
		return err
	}
	PtUtil.Debug("server start, host: %s", port)

	for {
		conn, err1 := listener.Accept()
		if err1 != nil {
			PtUtil.Debug("server Accept fail, err: ", err1)
			return err1
		}
		go server.runConn(conn)
	}

	return nil
}

//runConn 新连接处理函数
func (server *GxTCPServer) runConn(conn net.Conn) {
	gxConn := NewTCPConn()
	gxConn.Conn = conn
	gxConn.Connected = true
	gxConn.Remote = conn.RemoteAddr().String()

	//生成通讯需要的密钥
	// if gxConn.ServerKey() != nil {
	// 	gxConn.Conn.Close()
	// 	continue
	// }

	server.mutex.Lock()
	//分配ID，一般ID用三个字节保存就可以，第四个字节保留
	for {
		server.ID = server.Mask & (server.ID + 1)
		if _, ok := server.IDConns[server.ID]; !ok {
			break
		}
	}
	gxConn.ID = server.ID
	gxConn.M = "Cli"

	server.IDConns[server.ID] = gxConn
	server.addrConns[conn.RemoteAddr().String()] = gxConn
	server.mutex.Unlock()

	server.Nc(gxConn)  //回调到NewConn

	go gxConn.runHeartbeat() 

	for {
		msg, err := gxConn.Recv()
		if err != nil {
			server.closeConn(gxConn)
			return
		}

		if msg.GetCmd() != PtStatic.CmdHeartbeat {
			PtUtil.Debug("<<==== remote[%s:%s], info: %s", gxConn.M, gxConn.Remote, msg.String())
		}

		if cb, ok := server.clientCmds[msg.GetCmd()]; ok {
			//消息已经被注册
			err = cb(gxConn, msg)
			PtMsg.FreeMessage(msg)
			if err != nil {
				//回调返回值不为空，则关闭连接
				server.closeConn(gxConn)
				return
			}
			continue
		}

		//消息没有被注册
		err = server.Rm(gxConn, msg)
		PtMsg.FreeMessage(msg)
		if err != nil {
			server.closeConn(gxConn)
			return
		}
	}
}

func (server *GxTCPServer) closeConn(gxConn *GxTCPConn) {
	server.Dc(gxConn)
	server.mutex.Lock()
	delete(server.addrConns, gxConn.Remote)
	delete(server.IDConns, gxConn.ID)
	server.mutex.Unlock()
	gxConn.Toc <- 0xFFFF
	gxConn.Conn.Close()
}

//ConnectCount 返回当前连接数量
func (server *GxTCPServer) ConnectCount() int {
	server.mutex.Lock()
	defer server.mutex.Unlock()

	return len(server.IDConns)
}

//FindConnByRetome 根据连接地址返回连接指针
func (server *GxTCPServer) FindConnByRetome(retome string) *GxTCPConn {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	info, ok := server.addrConns[retome]
	if ok {
		return info
	}
	return nil
}

//FindConnByID 根据连接ID返回连接指针
func (server *GxTCPServer) FindConnByID(ID uint32) *GxTCPConn {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	info, ok := server.IDConns[ID]
	if ok {
		return info
	}
	return nil
}

//HeartbeatCallback 心跳回调
func HeartbeatCallback(conn *GxTCPConn, msg *PtMsg.GxMessage) error {
	conn.Toc <- 0
	msg.SetID(conn.ID)
	msg.SetRet(PtStatic.RetSucc)
	conn.Send(msg)
	return nil
}

//SendRawMessage 发送一个字符串消息
func SendRawMessage(conn *GxTCPConn, mask uint16, ID uint32, cmd uint16, seq uint16, ret uint16, buff []byte) {
	msg := PtMsg.GetGxMessage()
	defer PtMsg.FreeMessage(msg)

	msg.SetID(ID)
	msg.SetCmd(cmd)
	msg.SetRet(ret)
	msg.SetSeq(seq)
	msg.SetMask(mask)

	if len(buff) == 0 {
		msg.SetLen(0)
	} else {
		err := msg.Package(buff)
		if err != nil {
			PtUtil.Debug("PackagePbmsg error")
			return
		}
	}

	conn.Send(msg)
}

//SendPbMessage 发送一个pb消息
func SendPbMessage(conn *GxTCPConn, mask uint16, ID uint32, cmd uint16, seq uint16, ret uint16, pb proto.Message) {
	msg := PtMsg.GetGxMessage()
	defer PtMsg.FreeMessage(msg)

	msg.SetID(ID)
	msg.SetCmd(cmd)
	msg.SetRet(ret)
	msg.SetSeq(seq)
	msg.SetMask(mask)

	if pb == nil {
		msg.SetLen(0)
	} else {
		err := msg.PackagePbmsg(pb)
		if err != nil {
			PtUtil.Debug("PackagePbmsg error")
			return
		}
	}

	if msg.GetCmd() != PtStatic.CmdHeartbeat {
		if pb == nil {
			PtUtil.Debug("====>> remote[%s:%s], info: %s", conn.M, conn.Remote, msg.String())
		} else {
			PtUtil.Debug("====>> remote[%s:%s], info: %s, rsp: \r\n\t%s", conn.M, conn.Remote, msg.String(), pb.String())
		}
	}
	conn.Send(msg)
}

package PtNet

/**
 * Title:tcp连接接口
 * User: iuoon
 * Date: 2016-9-23
 * Version: 1.0
 */
import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/monnand/dhkx"
)

import (
	"strings"

	"github.com/iuoon/PuetxGo/PtMsg"
	"github.com/iuoon/PuetxGo/PtUtil"
)

//GxTCPConn tcp连接
type GxTCPConn struct {
	ID        uint32 //连接ID
	Conn      net.Conn
	ConnMutex sync.Mutex
	Connected bool //连接状态

	TimeoutCount int          //超时次数
	T            *time.Ticker //超时检测定时器
	Toc          chan int

	Data     []uint16 //支持的cmd列表
	ServerID uint32   //连接是服务器时，服务器ID
	M        string   //模块名 Cli，Adm，Srv
	Remote   string

	Key []byte //加密时使用的密钥

}

//NewTCPConn 生成一个新的GxTCPConn
func NewTCPConn() *GxTCPConn {
	tcpConn := new(GxTCPConn)
	tcpConn.Connected = false
	tcpConn.TimeoutCount = 0
	tcpConn.Toc = make(chan int, 1)
	tcpConn.T = time.NewTicker(5 * time.Second)
	tcpConn.M = "Cli" //默认

	tcpConn.ServerID = 0
	return tcpConn
}

//runHeartbeat 处理心跳函数，用协程启动
func (conn *GxTCPConn) runHeartbeat() {
	for {
		select {
		case state := <-conn.Toc:
			if state == 0XFFFF {
				return
			}
			conn.TimeoutCount = state
		case <-conn.T.C:
			if conn.TimeoutCount > 3 {
				//超时超过三次关闭连接
				conn.Conn.Close()

				PtUtil.Debug("client[%d] %s timeout", conn.ID, conn.Remote)
				return
			} else if conn.TimeoutCount >= 0 {
				conn.TimeoutCount = conn.TimeoutCount + 1
			} else {
				break
			}
		}
	}
}

//Send 发送消息函数
func (conn *GxTCPConn) Send(msg *PtMsg.GxMessage) error {
	conn.ConnMutex.Lock()
	defer conn.ConnMutex.Unlock()

	//写消息头
	len, err := conn.Conn.Write(msg.Header)
	if err != nil {
		PtUtil.Error("write header error:%s", err)
	}
	if uint16(len) != PtMsg.MessageHeaderLen {
		return errors.New("send error")
	}
	if err = msg.CheckFormat(); err != nil {
		return err
	}

	//如果消息体没有数据，直接返回
	if msg.GetLen() == 0 {
		return nil
	}

	//写消息体
	len, err = conn.Conn.Write(msg.Data)
	if err != nil {
		PtUtil.Error("write body error:%s", err)
	}
	if uint16(len) != msg.GetLen() {
		return errors.New("send error")
	}
	return nil
}

//Recv 接受消息函数
func (conn *GxTCPConn) Recv() (*PtMsg.GxMessage, error) {
	//写消息头
	//如果读取消息失败，消息要归还给消息池
	msg := PtMsg.GetGxMessage()
	leng, err := conn.Conn.Read(msg.Header) //这里返回时将换行符 \r\n 写入了结尾
	if err != nil {
		PtMsg.FreeMessage(msg)
		conn.Connected = false
		return nil, err
	}
	PtUtil.Debug("消息头长度：%d,头内容：%s", leng, strings.Replace(string(msg.Header), "\r\n", "", 2))
	if uint16(leng) != PtMsg.MessageHeaderLen {
		PtMsg.FreeMessage(msg)
		return nil, errors.New("recv error")
	}
	if err = msg.CheckFormat(); err != nil {
		PtUtil.Warn("recv error format message, remote: %s", conn.Remote)
		PtMsg.FreeMessage(msg)
		return nil, err
	}
	PtUtil.Debug("消息长度：%d", msg.GetLen())
	//消息头没有数据，则返回
	if msg.GetLen() == 0 {
		return msg, nil
	}

	//写消息体
	msg.InitData()
	leng, err = conn.Conn.Read(msg.Data)
	PtUtil.Debug("消息体长度：%d,消息体内容：%s", leng, strings.Replace(string(msg.Data), "\r\n", "", 2))
	if err != nil {
		PtMsg.FreeMessage(msg)
		conn.Connected = false
		return nil, err
	}
	if uint16(leng) != msg.GetLen() {
		PtMsg.FreeMessage(msg)
		return nil, errors.New("recv error")
	}
	return msg, nil
}

//Connect 连接指定host
func (conn *GxTCPConn) Connect(host string) error {
	c, err := net.Dial("tcp", host)

	if err != nil {
		return err
	}
	conn.Conn = c
	conn.Connected = true
	conn.Remote = c.RemoteAddr().String()
	return nil
}

//ServerKey 服务器生成一个简单的密钥，需要和客户端交互
func (conn *GxTCPConn) ServerKey() error {
	//给客户端发送19字节的随机字符串
	randomStr := PtUtil.GetRandomString(16)
	n, err := conn.Conn.Write([]byte(randomStr))
	if err != nil {
		PtUtil.Error("send random string fail, error:%s ", err)
		return err
	}
	PtUtil.Trace("send random string ok, len: %d", n)

	clientRandomStr := make([]byte, 16)
	n, err = conn.Conn.Read(clientRandomStr)
	if err != nil {
		PtUtil.Error("recv client random string fail, error: ", err)
		return err
	}
	PtUtil.Trace("recv client random string ok, len: %d, clientRandomStr: %s", n, clientRandomStr)

	h := md5.New()
	io.WriteString(h, "ruiyue")
	io.WriteString(h, randomStr)
	io.WriteString(h, string(clientRandomStr))
	conn.Key = []byte(fmt.Sprintf("%x", h.Sum(nil)))[:8]
	PtUtil.Trace("new crype key, len: %d, key: %s", len(conn.Key), conn.Key)

	// 读取客户端发送的加密随机字符串
	enStr := make([]byte, 24)
	n, err = conn.Conn.Read(enStr)
	if err != nil {
		PtUtil.Error("recv cleint encrypt random string, error: %s", err)
		return err
	}
	PtUtil.Trace("recv cleint encrypt random string ok, len: %d, keylen: %d, enStr: %x", n, len(enStr), enStr)

	//解密客户端发送的加密随机字符串
	deStr, _ := PtUtil.DesDecrypt(enStr)
	if randomStr != string(deStr) {
		PtUtil.Error("random string error, randomStr: %s, destr: %s", randomStr, deStr)
		return errors.New("crypt key error")
	}

	return nil
}

//ClientKey 客户端生成一个简单的密钥，需要和服务端交互
func (conn *GxTCPConn) ClientKey() error {
	serverRandomStr := make([]byte, 16)
	n, err := conn.Conn.Read(serverRandomStr)
	if err != nil {
		PtUtil.Error("recv server random string fail, error: %s", err)
		return err
	}
	PtUtil.Trace("recv server random string ok, len: %d, clientRandomStr: %s", n, serverRandomStr)

	//给客户端发送19字节的随机字符串
	randomStr := PtUtil.GetRandomString(16)
	n, err = conn.Conn.Write([]byte(randomStr))
	if err != nil {
		PtUtil.Error("send random string fail, error: %s", err)
		return err
	}
	PtUtil.Trace("send random string ok, len: %d", n)

	h := md5.New()
	io.WriteString(h, "ruiyue")
	io.WriteString(h, string(serverRandomStr))
	io.WriteString(h, randomStr)
	conn.Key = []byte(fmt.Sprintf("%x", h.Sum(nil)))[:8]
	PtUtil.Trace("new crype key, len: %d, key: %s", len(conn.Key), conn.Key)

	//加密随机字符串
	enStr, err1 := PtUtil.DesEncrypt([]byte(randomStr))
	if err1 != err {
		PtUtil.Error("encrype random fail, error: %s", err)
		return err1
	}

	// 发送加密随机字符串
	n, err = conn.Conn.Write(enStr)
	if err != nil {
		PtUtil.Error("send encrypt random string fail, error: %s", err)
		return err
	}

	return nil
}

//ServerDhKey 服务器生成一个DHkey交换算法的密钥，需要和客户端交互
func (conn *GxTCPConn) ServerDhKey() error {
	//给客户端发送19字节的随机字符串
	randomStr := PtUtil.GetRandomString(16)
	n, err := conn.Conn.Write([]byte(randomStr))
	if err != nil {
		PtUtil.Error("send random string fail, error: %s", err)
		return err
	}
	PtUtil.Trace("send random string ok, len: %d", n)

	g, _ := dhkx.GetGroup(0)
	priv, _ := g.GeneratePrivateKey(nil)
	pub := priv.Bytes()
	PtUtil.Trace("new private key ok, pubkeylen: %d", len(pub))

	// 接受客户端的DH公钥
	b := make([]byte, len(pub))
	n, err = conn.Conn.Read(b)
	if err != nil {
		PtUtil.Error("reav client public key, error: %s", err)
		return err
	}
	PtUtil.Trace("recv client public key ok, len: %d, keylen: %d", n, len(b))

	// 发送服务端的DH公钥到服务端
	n, err = conn.Conn.Write(pub)
	if err != nil {
		PtUtil.Error("send server public key fail, error: %s", err)
		return err
	}
	PtUtil.Trace("send server public key ok, len: %d, keylen: %d", n, len(pub))

	//获取加密公钥
	clientPub := dhkx.NewPublicKey(b)
	k, _ := g.ComputeKey(clientPub, priv)
	conn.Key = k.Bytes()
	PtUtil.Trace("new crype key, len: %d, key: %x", len(conn.Key), conn.Key)

	// 读取客户端发送的加密随机字符串
	enStr := make([]byte, 24)
	n, err = conn.Conn.Read(enStr)
	if err != nil {
		PtUtil.Error("recv cleint encrypt random string, error: %s", err)
		return err
	}
	PtUtil.Trace("recv cleint encrypt random string ok, len: %d, keylen: %d, enStr: %x", n, len(enStr), enStr)

	//解密客户端发送的加密随机字符串
	deStr, _ := PtUtil.DesDecrypt(enStr)
	if randomStr != string(deStr) {
		PtUtil.Error("random string error, randomStr: %s, destr: %s", randomStr, deStr)
		return errors.New("crypt key error")
	}

	return nil
}

//ClientDhKey 客户端生成一个DHkey交换算法的密钥，需要和服务端交互
func (conn *GxTCPConn) ClientDhKey() error {
	//接受服务端发送的随机字符串
	randomStr := make([]byte, 16)
	n, err := conn.Conn.Read(randomStr)
	if err != nil {
		PtUtil.Error("recv random string fail, error: %s", err)
		return err
	}
	PtUtil.Trace("recv random string ok, randomStr: %s", randomStr)

	g, _ := dhkx.GetGroup(0)
	priv, _ := g.GeneratePrivateKey(nil)
	pub := priv.Bytes()
	PtUtil.Trace("new private key ok, pubkeylen: %d", len(pub))

	// 发送客户端端的DH公钥到服务端
	n, err = conn.Conn.Write(pub)
	if err != nil {
		PtUtil.Error("send client public key fail, error: %s", err)
		return err
	}
	PtUtil.Trace("send client public key ok, len: %d, keylen: %d", n, len(pub))

	// 接受服务端的DH公钥
	b := make([]byte, len(pub))
	n, err = conn.Conn.Read(b)
	if err != nil {
		PtUtil.Error("reav server public key, error: %s", err)
		return err
	}
	PtUtil.Trace("recv server public key ok, len: %d, keylen: %d", n, len(b))

	//获取加密公钥
	servertPub := dhkx.NewPublicKey(b)
	k, _ := g.ComputeKey(servertPub, priv)
	conn.Key = k.Bytes()
	PtUtil.Trace("new crype key, len: %d, key: %x", len(conn.Key), conn.Key)

	//加密随机字符串
	enStr, err1 := PtUtil.DesEncrypt([]byte(randomStr))
	if err1 != err {
		PtUtil.Error("encrype random fail, error: %s", err)
		return err1
	}
	PtUtil.Trace("encrype random ok, enStr: %x", enStr)

	// 发送加密随机字符串
	n, err = conn.Conn.Write(enStr)
	if err != nil {
		PtUtil.Error("send encrypt random string fail, error: %s", err)
		return err
	}

	return nil
}

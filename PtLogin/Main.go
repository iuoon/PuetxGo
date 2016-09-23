package main

/**
 * Title:登陆服务
 * User: iuoon
 * Date: 2016-9-23
 * Version: 1.0
 */
import (
	"runtime"
	"github.com/iuoon/PuetxGo/PtUtil"
	"github.com/iuoon/PuetxGo/PtMsg"
	"time"
	"strconv"
	"errors"
	"github.com/iuoon/PuetxGo/PtDB"
	"github.com/iuoon/PuetxGo/PtStatic"
	"github.com/iuoon/PuetxGo/PtNet"
)

var server *PtNet.GxTCPServer

//NewConn 新客户端连接
func NewConn(conn *PtNet.GxTCPConn) {
	PtUtil.Debug("new connnect, remote: %s", conn.Remote)
}

//DisConn 客户端断开连接
func DisConn(conn *PtNet.GxTCPConn) {
	PtUtil.Debug("dis connnect, remote: %s", conn.Remote)
}

//NewMessage 不需要处理的消息,目前Login没有不处理的函数
func NewMessage(conn *PtNet.GxTCPConn, msg *PtMsg.GxMessage) error {
	PtUtil.Debug("new message, remote: %s", conn.Remote)
	conn.Send(msg)
	return errors.New("close")
}


func startServer() {
	server = PtNet.NewGxTCPServer(0x00FFFFFF, NewConn, DisConn, NewMessage, true)
	server.RegisterClientCmd(PtStatic.CmdLogin, login)
	server.RegisterClientCmd(PtStatic.CmdRegister, register)
	server.RegisterClientCmd(PtStatic.CmdGetGatesInfo, getGatesInfo)
	server.Start(":" + strconv.Itoa(config.Port))
}



func main() {
	PtUtil.InitLogger("login")
	err := LoadConfig()
	if err != nil {
		PtUtil.Debug("load config fail！")
		return
	}
	runtime.GOMAXPROCS(runtime.NumCPU()) //设置多核运行 go默认单核运行
	if config.MemoryPool > 0 {
		PtUtil.Info("open memory pool ok, init-size: %d", config.MemoryPool)
		PtUtil.OpenGxMemoryPool(config.MemoryPool)
	}
	if config.MessagePool > 0 {
		PtUtil.Info("open message pool ok, init-size: %d", config.MessagePool)
		PtMsg.OpenGxMessagePool(config.MessagePool)
	}

	if config.MemoryPool > 0 || config.MessagePool > 0 {
		go func() {
			t := time.NewTicker(3600 * time.Second)

			for {
				select {
				case <-t.C:
					PtUtil.PrintfMemoryPool()
					PtMsg.PrintfMessagePool()
				}
			}
		}()
	}

	err = PtDB.ConnectRedis(config.RedisHost, config.RedisPort, config.RedisDb)
	if err != nil {
		PtUtil.Debug("connect redis fail, err: %s", err)
		return
	}
	PtUtil.Debug("connect redis ok, host: %s:%d", config.RedisHost, config.RedisPort)
	err = PtDB.InitMysql(config.DbUser, config.DbPwd, config.DbHost, config.DbPort, config.DbDb, config.DbCharset)
	if err != nil {
		PtUtil.Debug("connect mysql fail, err: %s", err)
		return
	}
	PtUtil.Debug("connect mysql ok, host: %s:%d", config.DbHost, config.DbPort)

	err = LoadPlayer()
	if err != nil {
		PtUtil.Debug("load player fail, err: %s", err)
		return
	}
	PtUtil.Debug("load player ok")

	startServer()
}

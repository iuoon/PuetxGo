package main

/**
 * Title:消息管理
 * User: iuoon
 * Date: 2016-9-23
 * Version: 1.0
 */

import (
	"errors"
	"github.com/golang/protobuf/proto"
	"gopkg.in/redis.v3"
)

import (
	"git.oschina.net/jkkkls/goxiang/GxProto"
	"github.com/iuoon/PuetxGo/PtUtil"
	"github.com/iuoon/PuetxGo/PtDB"
	"github.com/iuoon/PuetxGo/PtStatic"
	"github.com/iuoon/PuetxGo/PtMsg"
	"github.com/iuoon/PuetxGo/PtNet"
)

func fillLoginRsp(rdClient *redis.Client, player *PtStatic.Player, rsp *GxProto.LoginServerRsp) {
	rsp.Info = &GxProto.LoginRspInfo{
		Token: proto.String(player.SaveToken(rdClient)),
	}

	var gates []*PtStatic.GateInfo
	PtStatic.GetAllGate(rdClient, &gates)
	var gate *PtStatic.GateInfo = nil
	for i := 0; i < len(gates); i++ {
		// if (time.Now().Unix() - gates[i].Ts) > 20 {
		// 	continue
		// }
		if gate == nil {
			gate = gates[i]
		} else {
			if gate.Count > gates[i].Count {
				gate = gates[i]
			}
		}
	}

	if gate != nil {
		rsp.GetInfo().Host = proto.String(gate.Host1)
		rsp.GetInfo().Port = proto.Uint32(uint32(gate.Port1))
	}

	serverID, ts := PtStatic.GetPlayerLastServer(rdClient, player.Username)
	var servers []*PtStatic.GameServer
	PtStatic.GetAllGameServer(rdClient, &servers)
	for i := 0; i < len(servers); i++ {
		PtUtil.Debug("server-ID: %d, name: %s", servers[i].ID, servers[i].Name)
		var lastts int64 = 0
		if serverID == servers[i].ID {
			lastts = ts
		}
		rsp.GetInfo().Srvs = append(rsp.GetInfo().Srvs, &GxProto.GameSrvInfo{
			Index:  proto.Uint32(uint32(servers[i].ID)),
			Name:   proto.String(servers[i].Name),
			Status: proto.Uint32(servers[i].Status),
			Lastts: proto.Uint32(uint32(lastts)),
		})
	}

}

func login(conn *PtNet.GxTCPConn, msg *PtMsg.GxMessage) error {
	rdClient := PtDB.PopRedisClient()
	defer PtDB.PushRedisClient(rdClient)

	var req GxProto.LoginServerReq
	var rsp GxProto.LoginServerRsp
	err := msg.UnpackagePbmsg(&req)
	if err != nil {
		PtUtil.Debug("UnpackagePbmsg error")
		return errors.New("close")
	}
	if req.GetRaw() == nil {
		PtUtil.Debug("login message miss filed: raw")
		PtNet.SendPbMessage(conn, 0, 0, msg.GetCmd(), msg.GetSeq(), PtStatic.RetMsgFormatError, nil)
		return errors.New("close")
	}

	PtUtil.Debug("login player, username: %s, pwd: %s", req.GetRaw().GetUsername(), req.GetRaw().GetPwd())

	player := FindPlayer(req.GetRaw().GetUsername())
	if player == nil {
		PtUtil.Debug("user is not exists, username: %s", req.GetRaw().GetUsername())
		PtNet.SendPbMessage(conn, 0, 0, msg.GetCmd(), msg.GetSeq(), PtStatic.RetUserNotExists, nil)
		return errors.New("close")
	}

	if !PtStatic.VerifyPassword(player, req.GetRaw().GetPwd()) {
		PtNet.SendPbMessage(conn, 0, 0, msg.GetCmd(), msg.GetSeq(), PtStatic.RetPwdError, nil)
		return errors.New("close")
	}

	PtUtil.Debug("old user: %s login from %s", req.GetRaw().GetUsername(), conn.Remote)
	fillLoginRsp(rdClient, player, &rsp)

	PtNet.SendPbMessage(conn, 0, 0, msg.GetCmd(), msg.GetSeq(), PtStatic.RetSucc, &rsp)
	return errors.New("close")
}

func register(conn *PtNet.GxTCPConn, msg *PtMsg.GxMessage) error {
	rdClient := PtDB.PopRedisClient()
	defer PtDB.PushRedisClient(rdClient)

	var req GxProto.LoginServerReq
	var rsp GxProto.LoginServerRsp
	err := msg.UnpackagePbmsg(&req)
	if err != nil {
		PtUtil.Debug("UnpackagePbmsg error")
		return errors.New("close")
	}

	if req.GetRaw() == nil {
		PtUtil.Debug("register message miss filed: raw")
		PtNet.SendPbMessage(conn, 0, 0, msg.GetCmd(), msg.GetSeq(), PtStatic.RetMsgFormatError, nil)
		return errors.New("close")
	}

	PtUtil.Debug("new player, username: %s, pwd: %s", req.GetRaw().GetUsername(), req.GetRaw().GetPwd())

	player := FindPlayer(req.GetRaw().GetUsername())
	if player != nil {
		PtUtil.Debug("user has been exists, username: %s", req.GetRaw().GetUsername())
		PtNet.SendPbMessage(conn, 0, 0, msg.GetCmd(), msg.GetSeq(), PtStatic.RetUserExists, nil)
		return errors.New("close")
	}

	player = PtStatic.NewPlayer(rdClient, req.GetRaw().GetUsername(), req.GetRaw().GetPwd(), uint32(req.GetPt()))
	err = AddPlayer(player)
	if err != nil {
		PtUtil.Debug("user has been exists, username: %s", req.GetRaw().GetUsername())
		PtNet.SendPbMessage(conn, 0, 0, msg.GetCmd(), msg.GetSeq(), PtStatic.RetUserExists, nil)
		return errors.New("close")
	}

	fillLoginRsp(rdClient, player, &rsp)

	PtUtil.Debug("new user: %s login from %s", req.GetRaw().GetUsername(), conn.Remote)
	PtNet.SendPbMessage(conn, 0, 0, msg.GetCmd(), msg.GetSeq(), PtStatic.RetSucc, &rsp)

	return errors.New("close")
}

func getGatesInfo(conn *PtNet.GxTCPConn, msg *PtMsg.GxMessage) error {
	rdClient := PtDB.PopRedisClient()
	defer PtDB.PushRedisClient(rdClient)

	var rsp GxProto.GetGatesInfoRsp
	var gates []*PtStatic.GateInfo
	PtStatic.GetAllGate(rdClient, &gates)
	for i := 0; i < len(gates); i++ {
		rsp.Host = append(rsp.Host, gates[i].Host1)
		rsp.Port = append(rsp.Port, uint32(gates[i].Port1))
	}
	PtNet.SendPbMessage(conn, 0, 0, msg.GetCmd(), msg.GetSeq(), PtStatic.RetSucc, &rsp)
	return errors.New("close")
}

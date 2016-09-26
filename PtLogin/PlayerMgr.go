package main

/**
 * Title:账户管理
 * User: iuoon
 * Date: 2016-9-23
 * Version: 1.0
 */

import (
	"container/list"
	"errors"
	"fmt"
	"sync"
	"time"
)

import (
	"github.com/iuoon/PuetxGo/PtDB"
	"github.com/iuoon/PuetxGo/PtStatic"
	"github.com/iuoon/PuetxGo/PtUtil"
)

//Players 帐号列表
type Players map[string]*PtStatic.Player

var players Players
var playersMutex *sync.Mutex

var sqlList *list.List
var sqlMutex *sync.Mutex

func init() {
	sqlList = list.New()
	sqlMutex = new(sync.Mutex)

	players = make(Players)
	playersMutex = new(sync.Mutex)
}

//PushPlayerSQL 保存一个帐号操作sql
func PushPlayerSQL(sql string) {
	sqlMutex.Lock()
	defer sqlMutex.Unlock()

	sqlList.PushBack(sql)
}

//PopPlayerSQL 取出一个帐号操作sql
func PopPlayerSQL() string {
	sqlMutex.Lock()
	defer sqlMutex.Unlock()

	if sqlList.Len() == 0 {
		return ""
	}

	sql := sqlList.Front().Value.(string)
	sqlList.Remove(sqlList.Front())
	return sql
}

//LoadPlayer 加载所有帐号
func LoadPlayer() error {
	str := PtDB.GenerateSelectAllSQL(&PtStatic.Player{}, "")
	PtUtil.Info("load all users sql:%s", str)
	rows, err := PtDB.MysqlPool.Query(str)
	defer rows.Close()
	if err != nil {
		PtUtil.Error("excsql err: %s", err)
		return err
	}

	n := 0
	for rows.Next() {
		player := new(PtStatic.Player)
		err = rows.Scan(&player.ID, &player.Username, &player.Password, &player.CreateTs, &player.Platform)
		players[player.Username] = player
		n++
	}
	PtUtil.Debug("load player, count: %d", n)

	go func() {
		for {
			sql := PopPlayerSQL()
			if sql == "" {
				time.Sleep(time.Second * 1)
				continue
			}
			_, err := PtDB.MysqlPool.Exec(sql)
			if err != nil {
				fmt.Println(err, ",sql:", sql)
			}
		}

	}()

	return nil
}

//FindPlayer 根据帐号名返回帐号信息
func FindPlayer(name string) *PtStatic.Player {
	playersMutex.Lock()
	defer playersMutex.Unlock()

	player, ok := players[name]
	if ok {
		return player
	}

	return nil
}

//AddPlayer 添加一个新帐号
func AddPlayer(player *PtStatic.Player) error {
	playersMutex.Lock()
	defer playersMutex.Unlock()

	_, ok := players[player.Username]
	if ok {
		return errors.New("player name is exists")
	}

	players[player.Username] = player
	PushPlayerSQL(PtDB.GenerateInsertSQL(player, ""))
	return nil
}

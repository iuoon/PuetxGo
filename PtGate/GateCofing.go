package main

/**
 * Title:加载配置
 * User: iuoon
 * Date: 2016-9-22
 * Version: 1.0
 */
import (
	"github.com/go-ini/ini"
	"github.com/iuoon/PuetxGo/PtUtil"
)

//GateConfig 网关配置
type GateConfig struct {
	Host1 string
	Port1 int
	Host2 string
	Port2 int

	DbHost    string
	DbPort    int
	DbUser    string
	DbPwd     string
	DbDb      string
	DbCharset string

	RedisHost string
	RedisPort int
	RedisDb   string
}

var config *GateConfig

func LoadConfig() error {

	cfg, err := ini.LoadSources(ini.LoadOptions{IgnoreContinuation: true}, "../Resource/gate.ini")
	if err != nil {
		return err
	}
	config = new(GateConfig)
	config.Host1 = cfg.Section("Server").Key("host1").String()
	config.Port1, _ = cfg.Section("Server").Key("port1").Int()
	config.Host2 = cfg.Section("Server").Key("host2").String()
	config.Port2, _ = cfg.Section("Server").Key("port2").Int()

	config.DbHost = cfg.Section("DB").Key("host").String()
	config.DbPort, _ = cfg.Section("DB").Key("port").Int()
	config.DbUser = cfg.Section("DB").Key("user").String()
	config.DbPwd = cfg.Section("DB").Key("pwd").String()
	config.DbDb = cfg.Section("DB").Key("db").String()
	config.DbCharset = cfg.Section("DB").Key("charset").String()

	config.RedisHost = cfg.Section("Redis").Key("host").String()
	config.RedisPort, _ = cfg.Section("Redis").Key("port").Int()
	config.RedisDb = cfg.Section("Redis").Key("db").String()

	PtUtil.Info("#################Puetx Config#######################")
	PtUtil.Info("Host1      : %s", config.Host1)
	PtUtil.Info("Port1      : %d", config.Port1)
	PtUtil.Info("Host2      : %s", config.Host2)
	PtUtil.Info("Port2      : %d", config.Port2)
	PtUtil.Info("DbHost     : %s", config.DbHost)
	PtUtil.Info("DbPort     : %d", config.DbPort)
	PtUtil.Info("DbUser     : %s", config.DbUser)
	PtUtil.Info("DbPwd      : %s", config.DbPwd)
	PtUtil.Info("DbDb       : %s", config.DbDb)
	PtUtil.Info("DbCharset  : %s", config.DbCharset)
	PtUtil.Info("RedisHost  : %s", config.RedisHost)
	PtUtil.Info("RedisPort  : %d", config.RedisPort)
	PtUtil.Info("RedisDb    : %s", config.RedisDb)
	PtUtil.Info("#####################################################")

	return nil

}

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
	Port int
	LogLevel int
	MemoryPool int
	MessagePool int

	DbHost    string
	DbPort    int
	DbUser    string
	DbPwd     string
	DbDb      string
	DbCharset string

	RedisHost string
	RedisPort int
	RedisDb   int64
}

var config *GateConfig

func LoadConfig() error {

	cfg, err := ini.LoadSources(ini.LoadOptions{IgnoreContinuation: true}, "../Resource/gate.ini")
	if err != nil {
		return err
	}
	config = new(GateConfig)
	config.Port,_ = cfg.Section("LoginServer").Key("port").Int()
	config.LogLevel, _ = cfg.Section("Server").Key("logLevel").Int()
	config.MemoryPool, _ = cfg.Section("Server").Key("memoryPool").Int()
	config.MessagePool, _ = cfg.Section("Server").Key("messagePool").Int()

	config.DbHost = cfg.Section("DB").Key("host").String()
	config.DbPort, _ = cfg.Section("DB").Key("port").Int()
	config.DbUser = cfg.Section("DB").Key("user").String()
	config.DbPwd = cfg.Section("DB").Key("pwd").String()
	config.DbDb = cfg.Section("DB").Key("db").String()
	config.DbCharset = cfg.Section("DB").Key("charset").String()

	config.RedisHost = cfg.Section("Redis").Key("host").String()
	config.RedisPort, _ = cfg.Section("Redis").Key("port").Int()
	config.RedisDb,_ = cfg.Section("Redis").Key("db").Int64()

	PtUtil.SetLevel(config.LogLevel)
	PtUtil.Info("#################Puetx Config#######################")
	PtUtil.Info("Port       : %d", config.Port)
	PtUtil.Info("LogLevel   : %d", config.LogLevel)
	PtUtil.Info("MemoryPool : %d", config.MemoryPool)
	PtUtil.Info("MessagePool: %d", config.MessagePool)
	PtUtil.Info("DbHost     : %s", config.DbHost)
	PtUtil.Info("DbPort     : %d", config.DbPort)
	PtUtil.Info("DbUser     : %s", config.DbUser)
	PtUtil.Info("DbPwd      : %s", config.DbPwd)
	PtUtil.Info("DbDb       : %s", config.DbDb)
	PtUtil.Info("DbCharset  : %s", config.DbCharset)
	PtUtil.Info("RedisHost  : %s", config.RedisHost)
	PtUtil.Info("RedisPort  : %d", config.RedisPort)
	PtUtil.Info("RedisDb    : %d", config.RedisDb)
	PtUtil.Info("#####################################################")

	return nil

}

package main

import (
	"log"

	"github.com/go-ini/ini"
)

func LoadConfig() bool {
	cfg, err := ini.LoadSources(ini.LoadOptions{IgnoreContinuation: true, AllowBooleanKeys: true}, "gate.ini")
	if err == nil {
		host := cfg.Section("Redis").Key("host").String()
		port := cfg.Section("Redis").Key("port").String()
		password := cfg.Section("Redis").Key("password").String()

		log.Println("#################redis#######################")
		log.Println("host:", host)
		log.Println("port:", port)
		log.Println("password:", password)
		return true
	} else {
		return false
	}

}

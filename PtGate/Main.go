package main

/**
 * Title:网关
 * User: iuoon
 * Date: 2016-9-22
 * Version: 1.0
 */
import (
	"log"
)

func main() {
	flag := LoadConfig()
	if !flag {
		log.Println("加载配置失败！")
		return
	} else {
		log.Println("redis配置加载成功！")
	}

}

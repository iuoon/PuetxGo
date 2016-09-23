package main

/**
 * Title:加载配置
 * User: iuoon
 * Date: 2016-9-22
 * Version: 1.0
 */
import (
	"github.com/iuoon/PuetxGo/PtUtil"
)

func main() {
	PtUtil.InitLogger("puetx")
	err := LoadConfig()
	if err != nil {
		PtUtil.Debug("load config fail！")
		return
	}

}

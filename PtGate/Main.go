package main

/**
 * Title:网关服务
 * User: iuoon
 * Date: 2016-9-22
 * Version: 1.0
 */
import (
	"github.com/iuoon/PuetxGo/PtUtil"
	"runtime"
)

func main() {
	PtUtil.InitLogger("gate")
	err := LoadConfig()
	if err != nil {
		PtUtil.Debug("load config fail！")
		return
	}
	runtime.GOMAXPROCS(runtime.NumCPU()) //设置多核运行 go默认单核运行
}

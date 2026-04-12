package main

import (
	"LittleTalk/core"
	"LittleTalk/flags"
	"LittleTalk/global"
)

func main() {
	flags.Parse()
	global.Config = core.ReadConf()
	core.InitLogrus()
	global.DB = core.InitDB()
	flags.Run()

}

package main

import (
	"LittleTalk/core"
	"LittleTalk/flags"
	"LittleTalk/global"
	"LittleTalk/router"
	"LittleTalk/service"
)

func main() {
	flags.Parse()

	global.Config = core.ReadConf()

	core.InitLogrus()
	global.RDB = core.InitRedis()
	global.DB = core.InitDB()
	flags.Run()

	service.Run()
	router.Run()
}

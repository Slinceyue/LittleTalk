package main

import (
	"LittleTalk/flags"
	"LittleTalk/global"
)

func main() {
	flags.Parse()
	global.Config = core.ReadConf()
}

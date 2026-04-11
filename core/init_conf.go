package core

import (
	"LittleTalk/conf"
	"LittleTalk/flags"
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
)

func ReadConf() (c *conf.Config) {
	bytedata, err := os.ReadFile(flags.FlagOptions.File)
	if err != nil {
		panic(err)
	}
	c = new(conf.Config)
	err = yaml.Unmarshal(bytedata, c)
	if err != nil {
		panic(fmt.Sprintf("配置文件格式错误: %s", err))
	}
	fmt.Printf("读取配置 %s 文件成功]\n", flags.FlagOptions.File)
	return
}

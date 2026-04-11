package flags

import (
	"LittleTalk/global"
	"LittleTalk/models"

	"github.com/sirupsen/logrus"
)

func FlagDB() {
	err := global.DB
	if err != nil {
		logrus.Errorf("数据库迁移失败 %s", err)
	}
	logrus.Infof("数据库迁移成功")
}

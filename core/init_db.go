package core

import (
	"LittleTalk/global"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

func InitDB() *gorm.DB {
	dc := global.Config.DB       // 从库（读操作）
	dcMaster := global.Config.DBMaster // 主库（写操作）
	db, err := gorm.Open(mysql.Open(dc.DSN()), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true, // 不生成外键约束
	})
	if err != nil {
		logrus.Fatalf("数据库连接失败: %s", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		logrus.Fatalf("获取数据库连接失败: %s", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	logrus.Infof("数据库连接成功")

	if !dcMaster.Empt() {
		// 主库配置不为空时注册读写分离
		err = db.Use(dbresolver.Register(dbresolver.Config{
			Sources:  []gorm.Dialector{mysql.Open(dcMaster.DSN())}, // 主库-写
			Replicas: []gorm.Dialector{mysql.Open(dc.DSN())},        // 从库-读
			Policy:   dbresolver.RandomPolicy{},                      // 负载均衡策略
		}))
		if err != nil {
			logrus.Fatalf("读写分离配置错误: %s", err)
		}
		logrus.Infof("读写分离配置已启用 [主库: %s:%d, 从库: %s:%d]",
			dcMaster.Host, dcMaster.Port, dc.Host, dc.Port)
	}

	return db
}

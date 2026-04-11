package core

import (
	"fmt"
	"strings"

	ipUtils "LittleTalk/utils/ip"
	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
	"github.com/sirupsen/logrus"
)

var searcher *xdb.Searcher

func InitIPDB() {
	dbath := "init/ip2region_v4.xdb"
	_searcher, err := xdb.NewWithFileOnly(xdb.IPv4, dbath)
	if err != nil {
		logrus.Fatal("ip数据库加载失败 %s", err)
		return
	}
	searcher = _searcher
}

func GetIPAddr(ip string) string {
	if ipUtils.HasLocalIPAddr(ip) {
		return "内网ip"
	}
	region, err := searcher.Search(ip)
	if err != nil {
		logrus.Warnf("错误的ip地址 %s", err)
		return "异常地址"
	}
	_addrList := strings.Split(region, "|")
	if len(_addrList) != 5 {
		logrus.Warnf("异常的的ip地址 %s", ip)
		return "未知地址"
	}
	country := _addrList[0]
	province := _addrList[1]
	city := _addrList[2]
	if province != "0" && city != "0" {
		return fmt.Sprintf("%s·%s", province, city)
	}
	if province != "0" && country != "0" {
		return fmt.Sprintf("%s·%s", country, province)
	}
	if country != "0" {
		return fmt.Sprintf("%s·%s", country, city)
	}
	return region
}

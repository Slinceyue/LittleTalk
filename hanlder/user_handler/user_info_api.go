package user_handler

import (
	"LittleTalk/models/enum"
	"time"
)

type UserInfo struct {
	Username  string            `json:"username"`
	Sex       enum.Sex          `json:"sex"`
	Avatar    string            `json:"avatar"`
	Intro     string            `json:"intro"`
	Birthday  string            `json:"birthday"`
	Status    enum.OnlineStatus `json:"status"`
	LastLogin time.Time         `json:"last_login"`
	IP        string            `json:"ip"`
}

func (UserHandler) UserInfo(userID uint) {

}

package hanlder

import (
	"LittleTalk/hanlder/friend_handler"
	"LittleTalk/hanlder/login_handler"
	"LittleTalk/hanlder/message_handler"
	"LittleTalk/hanlder/user_handler"
)

type AllApi struct {
	login_handler.LogApi
	message_handler.MessageApi
	friend_handler.FriendApi
	user_handler.UserApi
}

var Api AllApi

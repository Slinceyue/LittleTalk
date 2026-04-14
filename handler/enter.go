package handler

import (
	"LittleTalk/handler/friend_handler"
	"LittleTalk/handler/login_handler"
	"LittleTalk/handler/message_handler"
	"LittleTalk/handler/middle_handler"
	"LittleTalk/handler/user_handler"
)

type AllApi struct {
	login_handler.LoginHandler
	message_handler.MessageHandler
	friend_handler.FriendHandler
	user_handler.UserHandler
	middle_handler.MiddleHandler
}

var Api AllApi

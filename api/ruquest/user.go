package ruquest

import "LittleTalk/models/enum"

type NewUserRequest struct {
	Username string   `json:"username"`
	Password string   `json:"password"`
	Sex      enum.Sex `json:"sex"`
	Avatar   string   `json:"avatar"`
	Birthday string   `json:"birthday"`
}

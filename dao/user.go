package dao

import (
	"LittleTalk/api/request"
	"errors"
)

func CreatUser(user *request.NewUserRequest) error {
	if user == nil {
		err := errors.New("用户请求体异常")
		return err
	}
	if user.Username == "" || user.Password == "" {
		err := errors.New("用户名不能为空")
		return err
	}
	return nil
}

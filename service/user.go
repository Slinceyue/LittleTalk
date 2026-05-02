package service

import (
	"LittleTalk/api/request"
	"LittleTalk/api/response"
	"LittleTalk/cache"
	"LittleTalk/dao"
	"LittleTalk/models"
	"LittleTalk/utils/jwts"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
)

func CreatUser(ctx context.Context, request request.NewUserRequest) error {
	salt, err := genBackSalt()
	if err != nil {
		return errors.New("生成盐值异常")
	}
	pwd := calcFinalHash(request.Password, salt)
	return dao.CreatUser(ctx, models.User{
		Username: request.Username,
		Password: pwd,
		BackSalt: salt,
		Sex:      request.Sex,
		Birthday: request.Birthday,
	})
}
func LoginUser(ctx context.Context, loginRequest request.LoginRequest) (string, error) {
	logrus.Info("LoginUser called with username: ", loginRequest.Username, " userID: ", loginRequest.UserID)
	// 1. 参数校验（修复逻辑 ||）
	if loginRequest.Username == "" && loginRequest.UserID == 0 {
		return "", errors.New("用户名或ID不能为空")
	}
	if loginRequest.Password == "" {
		return "", errors.New("密码不能为空")
	}

	// 2. 查询用户（修复查询逻辑 + 作用域）
	var user models.User
	var err error

	switch {
	case loginRequest.Username != "":
		// 按用户名查询
		user, err = dao.GetByKey(ctx, models.User{}, "username", loginRequest.Username)
	case loginRequest.UserID != 0:
		// 按ID查询
		user, err = dao.GetByID(ctx, models.User{}, loginRequest.UserID)
	}

	// 3. 统一错误（安全：不区分不存在/密码错）
	if err != nil {
		return "", errors.New("用户名或密码错误")
	}

	// 4. 密码校验
	if user.Password != calcFinalHash(loginRequest.Password, user.BackSalt) {
		return "", errors.New("用户名或密码错误")
	}

	// 5. 生成 Token
	token, err := jwts.GetToken(jwts.Claims{
		UserID:   user.ID,
		UserName: user.Username,
		Role:     user.Role,
	})
	if err != nil {
		return "", errors.New("系统错误1")
	}

	err = cache.SetUserLoginStatus(ctx, user.ID, token)
	if err != nil {
		return "", errors.New("系统错误2")
	}

	return token, nil
}
func GetUser(ctx context.Context, userID uint) (response.SelfUserResponse, error) {
	user, err := dao.GetByID(ctx, models.User{}, userID)
	if err != nil {
		return response.SelfUserResponse{}, errors.New("用户不存在")
	}
	return response.SelfUserResponse{
		ID:       user.ID,
		Avatar:   fmt.Sprintf("static/avatar/%d.jpg", userID),
		Username: user.Username,
		Sex:      user.Sex,
		Intro:    user.Intro,
		Birthday: user.Birthday,
	}, nil
}
func GetOtherUser(ctx context.Context, userID uint) (response.OtherUserResponse, error) {
	user, err := dao.GetByID(ctx, models.User{}, userID)
	if err != nil {
		return response.OtherUserResponse{}, errors.New("用户不存在")
	}
	return response.OtherUserResponse{
		ID:       user.ID,
		Avatar:   fmt.Sprintf("static/avatar/%d.jpg", userID),
		Username: user.Username,
		Sex:      user.Sex,
		Intro:    user.Intro,
		Birthday: user.Birthday,
	}, nil
}
func UserInfoUpdate(ctx context.Context, id uint) {
	//TODO: 完善
}
func genBackSalt() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
func calcFinalHash(frontPwd string, backSalt string) string {
	h := sha256.New()
	h.Write([]byte(frontPwd + backSalt))
	return hex.EncodeToString(h.Sum(nil))
}

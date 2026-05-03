package service

import (
	"LittleTalk/api/request"
	"LittleTalk/api/response"
	"LittleTalk/cache"
	"LittleTalk/dao"
	"LittleTalk/models"
	"LittleTalk/models/enum"
	"LittleTalk/utils/jwts"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// LoginResult 登录结果
type LoginResult struct {
	Token  string
	UserID uint
}

// CreatUser 创建用户
func CreatUser(ctx context.Context, req request.NewUserRequest) error {
	salt, err := genBackSalt()
	if err != nil {
		return enum.CodeServerError
	}

	pwd := calcFinalHash(req.Password, salt)
	err = dao.CreatUser(ctx, models.User{
		Username: req.Username,
		Password: pwd,
		BackSalt: salt,
		Sex:      req.Sex,
		Birthday: req.Birthday,
	})
	if err != nil {
		if errors.Is(err, dao.ErrDuplicateEntry) {
			return enum.CodeUserAlreadyExist
		}
		return enum.CodeUserCreateFailed
	}
	return nil
}

// LoginUser 用户登录
func LoginUser(ctx context.Context, loginRequest request.LoginRequest) (LoginResult, error) {
	logrus.Info("[Login] 用户登录请求: username=", loginRequest.Username)

	// 1. 参数校验
	if loginRequest.Username == "" && loginRequest.UserID == 0 {
		return LoginResult{}, enum.CodeInvalidParam
	}
	if loginRequest.Password == "" {
		return LoginResult{}, enum.CodeInvalidParam
	}

	// 2. 查询用户
	var user models.User
	var err error

	switch {
	case loginRequest.Username != "":
		user, err = dao.GetByKey(ctx, models.User{}, "username", loginRequest.Username)
	case loginRequest.UserID != 0:
		user, err = dao.GetByID(ctx, models.User{}, loginRequest.UserID)
	}

	// 3. 统一错误处理（安全考虑：不区分用户不存在/密码错误）
	if err != nil {
		return LoginResult{}, enum.CodePasswordWrong
	}

	// 4. 密码校验
	if user.Password != calcFinalHash(loginRequest.Password, user.BackSalt) {
		return LoginResult{}, enum.CodePasswordWrong
	}

	// 5. 生成 Token
	token, err := jwts.GetToken(jwts.Claims{
		UserID:   user.ID,
		UserName: user.Username,
		Role:     user.Role,
	})
	if err != nil {
		logrus.Error("[Login] Token生成失败: userID=", user.ID, ", err=", err)
		return LoginResult{}, enum.CodeServerError
	}

	// 6. 保存登录状态
	err = cache.SetUserLoginStatus(ctx, user.ID, token)
	if err != nil {
		logrus.Error("[Login] 缓存Token失败: userID=", user.ID, ", err=", err)
		return LoginResult{}, enum.CodeServerError
	}

	logrus.Info("[Login] 用户登录成功: userID=", user.ID, ", username=", user.Username)
	return LoginResult{
		Token:  token,
		UserID: user.ID,
	}, nil
}

// GetUser 获取当前用户信息（优先从缓存获取）
func GetUser(ctx context.Context, userID uint) (response.SelfUserResponse, error) {
	// 先尝试从缓存获取
	cachedUser, err := cache.GetUserInfoCache(ctx, userID)
	if err == nil && cachedUser != nil {
		return *cachedUser, nil
	}

	// 缓存未命中，从数据库获取
	user, err := dao.GetByID(ctx, models.User{}, userID)
	if err != nil {
		return response.SelfUserResponse{}, enum.CodeUserNotFound
	}

	// 获取头像URL，如果没有则返回默认头像
	avatar := user.Avatar
	if avatar == "" {
		avatar = fmt.Sprintf("/static/avatar/%d.jpg", userID)
	}

	// 获取创建时间戳
	var createdAt int64
	if !user.CreatedAt.IsZero() {
		createdAt = user.CreatedAt.Unix()
	}

	userInfo := response.SelfUserResponse{
		ID:        user.ID,
		Avatar:    avatar,
		Username:  user.Username,
		Sex:       user.Sex,
		Intro:     user.Intro,
		Phone:     user.Phone,
		Email:     user.Email,
		Birthday:  user.Birthday,
		Status:    user.Status,
		CreatedAt: createdAt,
	}

	// 写入缓存（异步，不阻塞返回）
	go func() {
		cache.SetUserInfoCache(context.Background(), userID, &userInfo)
	}()

	return userInfo, nil
}

// GetOtherUser 获取其他用户信息（隐藏敏感信息）
func GetOtherUser(ctx context.Context, userID uint) (response.OtherUserResponse, error) {
	user, err := dao.GetByID(ctx, models.User{}, userID)
	if err != nil {
		return response.OtherUserResponse{}, enum.CodeUserNotFound
	}

	// 获取头像URL，如果没有则返回默认头像
	avatar := user.Avatar
	if avatar == "" {
		avatar = fmt.Sprintf("/static/avatar/%d.jpg", userID)
	}

	return response.OtherUserResponse{
		ID:       user.ID,
		Avatar:   avatar,
		Username: user.Username,
		Sex:      user.Sex,
		Intro:    user.Intro,
		Birthday: user.Birthday,
		// Phone 和 Email 不返回，保护隐私
	}, nil
}

// GetUserInfosByIDs 批量获取用户信息
func GetUserInfosByIDs(ctx context.Context, userIDs []uint) ([]response.OtherUserResponse, error) {
	if len(userIDs) == 0 {
		return []response.OtherUserResponse{}, nil
	}

	var users []models.User
	err := dao.GetByIDs(ctx, &users, userIDs)
	if err != nil {
		return nil, err
	}

	result := make([]response.OtherUserResponse, 0, len(users))
	for _, user := range users {
		avatar := user.Avatar
		if avatar == "" {
			avatar = fmt.Sprintf("/static/avatar/%d.jpg", user.ID)
		}
		result = append(result, response.OtherUserResponse{
			ID:       user.ID,
			Avatar:   avatar,
			Username: user.Username,
			Sex:      user.Sex,
			Intro:    user.Intro,
			Birthday: user.Birthday,
		})
	}
	return result, nil
}

// UserInfoUpdate 更新用户信息
func UserInfoUpdate(ctx context.Context, userID uint, req request.UserUpdateRequest) enum.ResCode {
	// 获取当前用户
	user, err := dao.GetByID(ctx, models.User{}, userID)
	if err != nil {
		return enum.CodeUserNotFound
	}

	// 记录是否有变更
	hasChange := false

	// 更新用户名（防重复检查）
	if req.Username != "" && req.Username != user.Username {
		// 检查用户名是否已被占用（排除自己）
		var existingUser models.User
		existingUser, err = dao.GetByKey(ctx, existingUser, "username", req.Username)
		if err == nil && existingUser.ID != userID {
			return enum.CodeUsernameAlreadyExist
		}
		user.Username = req.Username
		hasChange = true
	}

	// 更新性别
	if req.Sex >= 0 && req.Sex <= 2 && req.Sex != user.Sex {
		user.Sex = req.Sex
		hasChange = true
	}

	// 更新个人简介
	if req.Intro != user.Intro {
		user.Intro = req.Intro
		hasChange = true
	}

	// 更新手机号（防重复检查）
	if req.Phone != "" && req.Phone != user.Phone {
		// 检查手机号是否已被占用
		var existingUser models.User
		err = dao.GetByPhone(ctx, &existingUser, req.Phone)
		if err == nil && existingUser.ID != userID {
			return enum.CodePhoneAlreadyExist
		}
		user.Phone = req.Phone
		hasChange = true
	}

	// 更新邮箱（防重复检查）
	if req.Email != "" && req.Email != user.Email {
		// 检查邮箱是否已被占用
		var existingUser models.User
		err = dao.GetByEmail(ctx, &existingUser, req.Email)
		if err == nil && existingUser.ID != userID {
			return enum.CodeEmailAlreadyExist
		}
		user.Email = req.Email
		hasChange = true
	}

	// 更新生日
	if req.Birthday != "" && req.Birthday != user.Birthday {
		user.Birthday = req.Birthday
		hasChange = true
	}

	// 如果没有任何变更，直接返回成功
	if !hasChange {
		return enum.CodeSuccess
	}

	// 延迟双删：先删除缓存，再更新数据库，再延迟删除缓存
	// 第一次删除缓存
	cache.DelUserInfoCache(ctx, userID)

	// 保存更新到数据库
	err = dao.UpdateUser(ctx, user)
	if err != nil {
		if dao.IsDuplicateEntry(err) {
			// 根据具体错误类型判断
			logrus.Error("[UserInfoUpdate] 更新失败，数据冲突: userID=", userID, ", err=", err)
			return enum.CodeUserAlreadyExist
		}
		logrus.Error("[UserInfoUpdate] 更新用户信息失败: userID=", userID, ", err=", err)
		return enum.CodeServerError
	}

	// 延迟删除缓存（延迟双删策略）
	go func() {
		time.Sleep(100 * time.Millisecond)
		cache.DelUserInfoCache(context.Background(), userID)
	}()

	logrus.Info("[UserInfoUpdate] 用户信息更新成功: userID=", userID)
	return enum.CodeSuccess
}

// genBackSalt 生成随机盐值
func genBackSalt() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// calcFinalHash 计算最终密码哈希
func calcFinalHash(frontPwd string, backSalt string) string {
	h := sha256.New()
	h.Write([]byte(frontPwd + backSalt))
	return hex.EncodeToString(h.Sum(nil))
}

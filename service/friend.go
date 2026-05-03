package service

import (
	"LittleTalk/api/request"
	"LittleTalk/cache"
	"LittleTalk/dao"
	"LittleTalk/global"
	"LittleTalk/models"
	"LittleTalk/models/enum"
	"LittleTalk/utils/ws"
	"context"
	"errors"
	"fmt"
	"sync"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// FriendInfo 好友信息结构体
type FriendInfo struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
	Online   bool   `json:"online"`
}

// friendLockMap 分布式锁映射
var friendLockMap = struct {
	sync.RWMutex
	locks map[uint]*sync.Mutex
}{locks: make(map[uint]*sync.Mutex)}

// getFriendLock 获取用户的好友操作锁
func getFriendLock(userID uint) *sync.Mutex {
	friendLockMap.Lock()
	defer friendLockMap.Unlock()
	if friendLockMap.locks[userID] == nil {
		friendLockMap.locks[userID] = &sync.Mutex{}
	}
	return friendLockMap.locks[userID]
}

// GetFriendList 获取好友列表
func GetFriendList(ctx context.Context, userID uint) ([]FriendInfo, error) {
	friendIDs, err := cache.GetFriendList(ctx, userID)
	if err == nil && len(friendIDs) > 0 {
		friendInfos, err := getFriendDetailsByIDs(ctx, friendIDs)
		if err == nil {
			return friendInfos, nil
		}
	}

	friendModels, _, err := dao.GetFriendListByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("dao get friend list: %w", err)
	}
	if len(friendModels) == 0 {
		return []FriendInfo{}, nil
	}

	friendIDs = []uint{}
	for _, m := range friendModels {
		friendID := m.FriendID
		if m.FriendID == userID {
			friendID = m.UserID
		}
		friendIDs = append(friendIDs, friendID)
	}

	// 异步回写缓存
	go func() {
		if len(friendIDs) > 0 {
			_ = cache.SetFriendList(context.Background(), userID, friendIDs)
		}
	}()

	friendInfos, err := getFriendDetailsByIDs(ctx, friendIDs)
	if err != nil {
		return nil, fmt.Errorf("get friend details: %w", err)
	}
	return friendInfos, nil
}

// getFriendDetailsByIDs 根据ID列表批量获取好友详细信息
func getFriendDetailsByIDs(ctx context.Context, friendIDs []uint) ([]FriendInfo, error) {
	if len(friendIDs) == 0 {
		return []FriendInfo{}, nil
	}

	var friends []models.User
	err := global.DB.WithContext(ctx).Where("id IN ?", friendIDs).Find(&friends).Error
	if err != nil {
		return nil, fmt.Errorf("query friends: %w", err)
	}

	// 直接查询内存WS连接状态，不再依赖Redis
	onlineMap := make(map[uint]bool, len(friendIDs))
	for _, id := range friendIDs {
		onlineMap[id] = ws.ConnManager.IsOnline(id)
	}

	result := make([]FriendInfo, 0, len(friends))
	for _, f := range friends {
		// 获取头像URL
		avatar := f.Avatar
		if avatar == "" {
			avatar = fmt.Sprintf("/static/avatar/%d.jpg", f.ID)
		}
		result = append(result, FriendInfo{
			ID:       f.ID,
			Username: f.Username,
			Avatar:   avatar,
			Online:   onlineMap[f.ID],
		})
	}
	return result, nil
}

// FriendRequest 发送好友请求
func FriendRequest(ctx context.Context, serviceRequest request.FriendRequest, userID uint) error {
	// 参数校验
	if serviceRequest.FriendID == 0 {
		return enum.CodeInvalidParam
	}
	if serviceRequest.FriendID == userID {
		return enum.CodeFriendSelfRequest
	}

	lock := getFriendLock(serviceRequest.FriendID)
	lock.Lock()
	defer lock.Unlock()

	// 检查目标用户是否存在
	_, err := dao.GetByID(ctx, models.User{}, serviceRequest.FriendID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return enum.CodeUserNotFound
		}
		return fmt.Errorf("check user: %w", err)
	}

	tx := global.DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("db error: %w", tx.Error)
	}

	// 检查是否已是好友
	var existingFriend models.Friend
	err = tx.Where("user_id = ? AND friend_id = ?", userID, serviceRequest.FriendID).
		Or("user_id = ? AND friend_id = ?", serviceRequest.FriendID, userID).
		First(&existingFriend).Error
	if err == nil {
		tx.Rollback()
		return enum.CodeFriendAlreadyExist
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		return fmt.Errorf("db error: %w", err)
	}

	// 检查是否已有待处理的申请
	var existingRequest models.FriendRequest
	err = tx.Where("from_user_id = ? AND to_user_id = ? AND status = 0", userID, serviceRequest.FriendID).
		First(&existingRequest).Error
	if err == nil {
		tx.Rollback()
		return enum.CodeFriendRequestExist
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		return fmt.Errorf("db error: %w", err)
	}

	// 创建好友申请记录
	err = tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&models.FriendRequest{
		FromUserID: userID,
		ToUserID:   serviceRequest.FriendID,
		Status:     0,
	}).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("db error: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("db error: %w", err)
	}

	// 异步通知目标用户
	go func() {
		bgCtx := context.Background()

		// 获取发送者用户名
		var fromUsername string
		var sender models.User
		if err := global.DB.WithContext(bgCtx).Where("id = ?", userID).First(&sender).Error; err == nil {
			fromUsername = sender.Username
		} else {
			fromUsername = fmt.Sprintf("用户%d", userID)
		}

		if IsUserOnline(serviceRequest.FriendID) {
			MessageChannel <- &request.MessageContext{
				MsgType: "friend",
				FriendMessageRequest: request.FriendMessageRequest{
					FromID:       userID,
					ToID:         serviceRequest.FriendID,
					FromUsername: fromUsername,
				},
			}
		}
		_ = cache.SaveOfflineMessage(bgCtx, serviceRequest.FriendID, serviceRequest)
		_, _ = cache.SetFriendRequest(bgCtx, serviceRequest.FriendID, userID)
	}()

	return nil
}

// GetFriendRequest 获取待处理的好友请求列表
func GetFriendRequest(ctx context.Context, userID uint) ([]uint, error) {
	ids, err := cache.GetFriendRequest(ctx, userID)
	if err == nil && len(ids) > 0 {
		return ids, nil
	}

	reqs, _, err := dao.ListQuery(ctx, models.FriendRequest{
		ToUserID: userID,
		Status:   0,
	}, dao.Options{})
	if err != nil {
		return nil, err
	}

	fromIDs := make([]uint, 0, len(reqs))
	pipe := global.RDB.Pipeline()
	key := fmt.Sprintf("friend:request:%d", userID)

	for _, r := range reqs {
		fromIDs = append(fromIDs, r.FromUserID)
		pipe.SAdd(ctx, key, r.FromUserID)
	}

	if len(reqs) > 0 {
		pipe.Expire(ctx, key, cache.ExpireStatus())
		_, err = pipe.Exec(ctx)
		if err != nil {
			return fromIDs, nil // 缓存回写失败不影响返回
		}
	}

	return fromIDs, nil
}

// OKFriendRequest 同意好友请求
func OKFriendRequest(ctx context.Context, request request.FriendRequestOK, userID uint) error {
	lock := getFriendLock(userID)
	lock.Lock()
	defer lock.Unlock()

	tx := global.DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("db error: %w", tx.Error)
	}

	// 检查申请是否存在
	var friendRequest models.FriendRequest
	if err := tx.Where("from_user_id = ? AND to_user_id = ? AND status = 0", request.FromID, userID).
		First(&friendRequest).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			return enum.CodeFriendRequestNotFound
		}
		tx.Rollback()
		return fmt.Errorf("db error: %w", err)
	}

	var friend models.Friend
	hasForward := false
	if err := tx.Where("user_id = ? AND friend_id = ?", userID, request.FromID).First(&friend).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			return fmt.Errorf("db error: %w", err)
		}
	} else {
		hasForward = true
	}

	hasReverse := false
	if err := tx.Where("user_id = ? AND friend_id = ?", request.FromID, userID).First(&friend).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			return fmt.Errorf("db error: %w", err)
		}
	} else {
		hasReverse = true
	}

	// 创建好友关系（双向）
	if !hasForward {
		err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&models.Friend{
			UserID:   userID,
			FriendID: request.FromID,
		}).Error
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("db error: %w", err)
		}
	}

	if !hasReverse {
		err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&models.Friend{
			UserID:   request.FromID,
			FriendID: userID,
		}).Error
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("db error: %w", err)
		}
	}

	// 更新申请状态为已同意
	err := tx.Model(&models.FriendRequest{}).
		Where("from_user_id = ? AND to_user_id = ?", request.FromID, userID).
		Update("status", 1).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("db error: %w", err)
	}

	// 删除申请记录
	err = tx.Where("from_user_id = ? AND to_user_id = ?", request.FromID, userID).
		Delete(&models.FriendRequest{}).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("db error: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("db error: %w", err)
	}

	// 异步清理缓存
	go func() {
		bgCtx := context.Background()
		_ = cache.DelFriendList(bgCtx, userID)
		_ = cache.DelFriendList(bgCtx, request.FromID)
		_ = cache.DelFriendRequest(bgCtx, userID, request.FromID)
	}()

	return nil
}

// RejectFriendRequest 拒绝好友请求
func RejectFriendRequest(ctx context.Context, fromID, userID uint) error {
	lock := getFriendLock(userID)
	lock.Lock()
	defer lock.Unlock()

	tx := global.DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("db error: %w", tx.Error)
	}

	// 检查申请是否存在
	var friendRequest models.FriendRequest
	if err := tx.Where("from_user_id = ? AND to_user_id = ? AND status = 0", fromID, userID).
		First(&friendRequest).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			return enum.CodeFriendRequestNotFound
		}
		tx.Rollback()
		return fmt.Errorf("db error: %w", err)
	}

	// 删除好友请求记录
	err := tx.Where("from_user_id = ? AND to_user_id = ?", fromID, userID).
		Delete(&models.FriendRequest{}).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("db error: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("db error: %w", tx.Error)
	}

	go func() {
		bgCtx := context.Background()
		_ = cache.DelFriendRequest(bgCtx, userID, fromID)
	}()

	return nil
}

// DeleteFriend 删除好友（双向删除）
func DeleteFriend(ctx context.Context, friendID, userID uint) error {
	if friendID == 0 {
		return enum.CodeInvalidParam
	}
	if friendID == userID {
		return enum.CodeFriendSelfRequest
	}

	lock := getFriendLock(userID)
	lock.Lock()
	defer lock.Unlock()

	tx := global.DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("db error: %w", tx.Error)
	}

	// 检查好友关系是否存在
	var friend models.Friend
	if err := tx.Where("(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)",
		userID, friendID, friendID, userID).
		First(&friend).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			return enum.CodeFriendNotFound
		}
		tx.Rollback()
		return fmt.Errorf("db error: %w", err)
	}

	// 删除双向好友关系
	err := tx.Where("(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)",
		userID, friendID, friendID, userID).
		Delete(&models.Friend{}).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("db error: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("db error: %w", err)
	}

	go func() {
		bgCtx := context.Background()
		_ = cache.DelFriendList(bgCtx, userID)
		_ = cache.DelFriendList(bgCtx, friendID)
	}()

	return nil
}

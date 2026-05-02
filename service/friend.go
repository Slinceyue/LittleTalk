package service

import (
	"LittleTalk/api/request"
	"LittleTalk/cache"
	"LittleTalk/dao"
	"LittleTalk/global"
	"LittleTalk/models"
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func GetFriendList(ctx context.Context, userID uint) ([]uint, error) {
	friends, err := cache.GetFriendList(ctx, userID)
	if err == nil {
		if len(friends) > 0 {
			return friends, nil
		}
		// 缓存可能是空集或过期后的残留，继续查数据库
	}
	// 3. 缓存未命中 → 查数据库（兜底）
	friendModels, _, err := dao.GetFriendListByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("dao get friend list: %w", err)
	}
	if len(friendModels) == 0 {
		return []uint{}, nil
	}
	// 4. 提取好友ID
	var friendIDs []uint
	for _, m := range friendModels {
		// 假设双向好友：当前user是用户A，则好友是用户B
		friendID := m.FriendID // 按你表结构改
		if m.FriendID == userID {
			friendID = m.UserID
		}
		friendIDs = append(friendIDs, friendID)
	}

	// 5. 回写缓存（批量SADD）
	if len(friendIDs) > 0 {
		err = cache.SetFriendList(ctx, userID, friendIDs)
		if err != nil {
			return nil, fmt.Errorf("cache set friend list: %w", err)
		}
	}

	return friendIDs, nil
}
func FriendRequest(ctx context.Context, serviceRequest request.FriendRequest, userID uint) error {
	if serviceRequest.FriendID == 0 || serviceRequest.FriendID == userID {
		return fmt.Errorf("invalid friend id")
	}
	// 0. 校验对方用户是否存在
	_, err := dao.GetByID(ctx, models.User{}, serviceRequest.FriendID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}
	// 0.1 校验是否已经是好友
	_, err = dao.Get(ctx, models.Friend{UserID: userID, FriendID: serviceRequest.FriendID})
	if err == nil {
		return fmt.Errorf("already friends")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("db error: %w", err)
	}
	// 0.2 校验是否已经发送过好友请求
	_, err = dao.Get(ctx, models.FriendRequest{FromUserID: userID, ToUserID: serviceRequest.FriendID})
	if err == nil {
		return fmt.Errorf("friend request already exists")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("db error: %w", err)
	}
	// 1. 入库
	err = dao.Creat(ctx, models.FriendRequest{
		FromUserID: userID,
		ToUserID:   serviceRequest.FriendID,
		Status:     0,
	})
	if err != nil {
		return fmt.Errorf("db error: %w", err)
	}
	ok := IsUserOnline(serviceRequest.FriendID)

	if ok {
		MessageChannel <- &request.MessageContext{
			MsgType: "friend",
			FriendMessageRequest: request.FriendMessageRequest{
				FromID: userID,
				ToID:   serviceRequest.FriendID,
			},
		}
		_ = cache.SaveOfflineMessage(ctx, serviceRequest.FriendID, serviceRequest)
	}
	//2. 存缓存：toUserID=对方，fromUserID=我
	_, _ = cache.SetFriendRequest(ctx, serviceRequest.FriendID, userID)
	return nil
}
func GetFriendRequest(ctx context.Context, userID uint) ([]uint, error) {
	fmt.Println("进入函数")
	// 1. 查缓存
	ids, err := cache.GetFriendRequest(ctx, userID)
	if err == nil && len(ids) > 0 {
		fmt.Println("缓存命中")
		return ids, nil
	}

	// 2. 缓存未命中 → 查库（只查待处理）
	reqs, _, err := dao.ListQuery(ctx, models.FriendRequest{
		ToUserID: userID,
		Status:   0,
	}, dao.Options{})
	fmt.Println("查库")
	if err != nil {
		return nil, err
	}
	fmt.Printf("查询结果：%v", reqs)
	// 3. 回写缓存
	var fromIDs []uint
	for _, r := range reqs {
		fromIDs = append(fromIDs, r.FromUserID)
		_, _ = cache.SetFriendRequest(ctx, userID, r.FromUserID)
		fmt.Println("回写")
	}

	return fromIDs, nil
}
func OKFriendRequest(ctx context.Context, request request.FriendRequestOK, userID uint) error {
	// 开启事务（必须！）
	tx := global.DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("db error: %w", tx.Error)
	}

	var err error
	// 1. 检查是否已经是好友，避免事务中重复插入
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

	if !hasForward {
		err = tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&models.Friend{
			UserID:   userID,
			FriendID: request.FromID,
		}).Error
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("db error: %w", err)
		}
	}

	if !hasReverse {
		err = tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&models.Friend{
			UserID:   request.FromID,
			FriendID: userID,
		}).Error
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("db error: %w", err)
		}
	}

	// 3. 更新好友申请状态（正确条件：from + to）
	err = tx.Model(&models.FriendRequest{}).
		Where("from_user_id = ? AND to_user_id = ?", request.FromID, userID).
		Update("status", 1).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("db error: %w", err)
	}

	// 4. 删除好友申请记录
	err = tx.Where("from_user_id = ? AND to_user_id = ?", request.FromID, userID).
		Delete(&models.FriendRequest{}).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("db error: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("db error: %w", err)
	}

	// 清理缓存
	_ = cache.DelFriendList(ctx, userID)
	_ = cache.DelFriendList(ctx, request.FromID)
	_ = cache.DelFriendRequest(ctx, userID, request.FromID)

	return nil
}

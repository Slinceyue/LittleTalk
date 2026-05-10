package dao

import (
	"LittleTalk/global"
	"LittleTalk/models"
	"context"
)

func GetFriendListByUserID(ctx context.Context, userID uint) (list []models.Friend, count int, err error) {
	return ListQuery(ctx, models.Friend{}, Options{
		Where: global.DB.Where("user_id = ? OR friend_id = ?", userID, userID),
	})
}

// IsFriend 检查两个用户是否为好友关系
func IsFriend(ctx context.Context, userID, targetID uint) (bool, error) {
	var count int64
	err := global.DB.Model(&models.Friend{}).
		Where("(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)",
			userID, targetID, targetID, userID).
		Limit(1).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

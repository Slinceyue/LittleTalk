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

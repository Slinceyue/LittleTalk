package dao

import (
	"LittleTalk/global"
	"LittleTalk/models"
	"context"
)

// CreateRoom 创建房间
func CreateRoom(ctx context.Context, room *models.Room) error {
	return global.DB.WithContext(ctx).Create(room).Error
}

// GetRoomByID 获取房间详情
func GetRoomByID(ctx context.Context, roomID uint) (*models.Room, error) {
	var room models.Room
	err := global.DB.WithContext(ctx).First(&room, roomID).Error
	if err != nil {
		return nil, err
	}
	return &room, nil
}

// UpdateRoom 更新房间信息
func UpdateRoom(ctx context.Context, room *models.Room) error {
	return global.DB.WithContext(ctx).Save(room).Error
}

// DeleteRoom 删除房间
func DeleteRoom(ctx context.Context, roomID uint) error {
	return global.DB.WithContext(ctx).Delete(&models.Room{}, roomID).Error
}

// GetUserRooms 获取用户所在的所有群
func GetUserRooms(ctx context.Context, userID uint) ([]models.Room, error) {
	var roomUsers []models.RoomUser
	err := global.DB.WithContext(ctx).Where("user_id = ?", userID).Find(&roomUsers).Error
	if err != nil {
		return nil, err
	}

	if len(roomUsers) == 0 {
		return []models.Room{}, nil
	}

	roomIDs := make([]uint, len(roomUsers))
	for i, ru := range roomUsers {
		roomIDs[i] = ru.RoomID
	}

	var rooms []models.Room
	err = global.DB.WithContext(ctx).Where("id IN ?", roomIDs).Find(&rooms).Error
	return rooms, err
}

// AddRoomMember 添加群成员
func AddRoomMember(ctx context.Context, roomID, userID uint, role int8) error {
	ru := models.RoomUser{
		RoomID: roomID,
		UserID: userID,
		Role:   role,
	}
	return global.DB.WithContext(ctx).Create(&ru).Error
}

// RemoveRoomMember 移除群成员
func RemoveRoomMember(ctx context.Context, roomID, userID uint) error {
	return global.DB.WithContext(ctx).Where("room_id = ? AND user_id = ?", roomID, userID).Delete(&models.RoomUser{}).Error
}

// GetRoomMembers 获取群成员列表
func GetRoomMembers(ctx context.Context, roomID uint) ([]models.RoomUser, error) {
	var members []models.RoomUser
	err := global.DB.WithContext(ctx).Where("room_id = ?", roomID).Find(&members).Error
	return members, err
}

// GetRoomMember 获取群成员
func GetRoomMember(ctx context.Context, roomID, userID uint) (*models.RoomUser, error) {
	var member models.RoomUser
	err := global.DB.WithContext(ctx).Where("room_id = ? AND user_id = ?", roomID, userID).First(&member).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

// UpdateRoomMember 更新群成员信息
func UpdateRoomMember(ctx context.Context, member *models.RoomUser) error {
	return global.DB.WithContext(ctx).Save(member).Error
}

// IsRoomMember 检查是否为群成员
func IsRoomMember(ctx context.Context, roomID, userID uint) bool {
	var count int64
	global.DB.WithContext(ctx).Model(&models.RoomUser{}).Where("room_id = ? AND user_id = ?", roomID, userID).Count(&count)
	return count > 0
}

// GetRoomMemberCount 获取群成员数量
func GetRoomMemberCount(ctx context.Context, roomID uint) int64 {
	var count int64
	global.DB.WithContext(ctx).Model(&models.RoomUser{}).Where("room_id = ?", roomID).Count(&count)
	return count
}

// UpdateRoomMemberCount 更新群成员数量
func UpdateRoomMemberCount(ctx context.Context, roomID uint) error {
	count := GetRoomMemberCount(ctx, roomID)
	return global.DB.WithContext(ctx).Model(&models.Room{}).Where("id = ?", roomID).Update("member_cnt", count).Error
}

// DeleteRoomAllMembers 删除群的所有成员
func DeleteRoomAllMembers(ctx context.Context, roomID uint) error {
	return global.DB.WithContext(ctx).Where("room_id = ?", roomID).Delete(&models.RoomUser{}).Error
}

// GetAllRooms 获取所有公开群（用于搜索）
func GetAllRooms(ctx context.Context, limit int) ([]models.Room, error) {
	var rooms []models.Room
	err := global.DB.WithContext(ctx).Order("created_at DESC").Limit(limit).Find(&rooms).Error
	return rooms, err
}

// SearchRooms 搜索群
func SearchRooms(ctx context.Context, keyword string) ([]models.Room, error) {
	var rooms []models.Room
	err := global.DB.WithContext(ctx).Where("name LIKE ?", "%"+keyword+"%").Limit(50).Find(&rooms).Error
	return rooms, err
}

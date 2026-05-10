package service

import (
	"LittleTalk/cache"
	"LittleTalk/dao"
	"LittleTalk/global"
	"LittleTalk/models"
	"LittleTalk/models/enum"
	"LittleTalk/utils/ws"
	"context"
	"fmt"
	"sync"
	"time"

	"gorm.io/gorm"
)

// RoomInfo 群信息
type RoomInfo struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	Avatar     string `json:"avatar"`
	Intro      string `json:"intro"`
	OwnerID    uint   `json:"owner_id"`
	MemberCnt  int    `json:"member_cnt"`
	CreatedAt  int64  `json:"created_at"`
}

// RoomMemberInfo 群成员信息
type RoomMemberInfo struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
	Role     int8   `json:"role"`
	Nickname string `json:"nickname"`
	JoinTime int64  `json:"join_time"`
	Online   bool   `json:"online"`
}

// GroupMessage 群消息
type GroupMessage struct {
	FromID      uint   `json:"from_id"`
	FromName    string `json:"from_name"`
	FromAvatar  string `json:"from_avatar"`
	RoomID      uint   `json:"room_id"`
	Content     string `json:"content"`
	MessageType int    `json:"message_type"`
	SendTime    int64  `json:"send_time"`
}

// roomLockMap 分布式锁映射
var roomLockMap = struct {
	sync.RWMutex
	locks map[uint]*sync.RWMutex
}{
	locks: make(map[uint]*sync.RWMutex),
}

func getRoomLock(roomID uint) *sync.RWMutex {
	roomLockMap.Lock()
	defer roomLockMap.Unlock()
	if _, exists := roomLockMap.locks[roomID]; !exists {
		roomLockMap.locks[roomID] = &sync.RWMutex{}
	}
	return roomLockMap.locks[roomID]
}

// CreateRoom 创建群聊
func CreateRoom(ctx context.Context, name string, userID uint) (*RoomInfo, error) {
	if name == "" {
		return nil, enum.CodeRoomNameEmpty
	}

	// 创建群
	room := &models.Room{
		Name:    name,
		OwnerID: userID,
	}

	err := dao.CreateRoom(ctx, room)
	if err != nil {
		return nil, err
	}

	// 创建者自动加入群
	err = dao.AddRoomMember(ctx, room.ID, userID, int8(enum.RoomRoleOwner))
	if err != nil {
		return nil, err
	}

	// 更新成员数量
	_ = dao.UpdateRoomMemberCount(ctx, room.ID)

	return &RoomInfo{
		ID:        room.ID,
		Name:      room.Name,
		Avatar:    room.Avatar,
		Intro:     room.Intro,
		OwnerID:   room.OwnerID,
		MemberCnt: 1,
		CreatedAt: time.Now().Unix(),
	}, nil
}

// GetRoomInfo 获取群信息
func GetRoomInfo(ctx context.Context, roomID uint) (*RoomInfo, error) {
	room, err := dao.GetRoomByID(ctx, roomID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, enum.CodeRoomNotFound
		}
		return nil, err
	}

	avatar := room.Avatar
	if avatar == "" {
		avatar = fmt.Sprintf("/static/room/%d.png", room.ID)
	}

	return &RoomInfo{
		ID:         room.ID,
		Name:       room.Name,
		Avatar:     avatar,
		Intro:      room.Intro,
		OwnerID:    room.OwnerID,
		MemberCnt:  room.MemberCnt,
		CreatedAt:  room.CreatedAt.Unix(),
	}, nil
}

// GetUserRooms 获取用户的群列表
func GetUserRooms(ctx context.Context, userID uint) ([]RoomInfo, error) {
	rooms, err := dao.GetUserRooms(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]RoomInfo, 0, len(rooms))
	for _, room := range rooms {
		avatar := room.Avatar
		if avatar == "" {
			avatar = fmt.Sprintf("/static/room/%d.png", room.ID)
		}
		result = append(result, RoomInfo{
			ID:         room.ID,
			Name:       room.Name,
			Avatar:     avatar,
			Intro:      room.Intro,
			OwnerID:    room.OwnerID,
			MemberCnt:  room.MemberCnt,
			CreatedAt:  room.CreatedAt.Unix(),
		})
	}

	return result, nil
}

// GetRoomMembers 获取群成员列表
func GetRoomMembers(ctx context.Context, roomID uint) ([]RoomMemberInfo, error) {
	members, err := dao.GetRoomMembers(ctx, roomID)
	if err != nil {
		return nil, err
	}

	// 收集所有用户ID
	userIDs := make([]uint, len(members))
	for i, m := range members {
		userIDs[i] = m.UserID
	}

	// 批量查询用户信息
	var users []models.User
	if len(userIDs) > 0 {
		err = global.DB.WithContext(ctx).Where("id IN ?", userIDs).Find(&users).Error
		if err != nil {
			return nil, err
		}
	}

	userMap := make(map[uint]models.User)
	for _, u := range users {
		userMap[u.ID] = u
	}

	// 构建成员信息
	result := make([]RoomMemberInfo, 0, len(members))
	for _, m := range members {
		user, ok := userMap[m.UserID]
		if !ok {
			continue
		}

		avatar := user.Avatar
		if avatar == "" {
			avatar = fmt.Sprintf("/static/avatar/%d.jpg", user.ID)
		}

		result = append(result, RoomMemberInfo{
			ID:       user.ID,
			Username: user.Username,
			Avatar:   avatar,
			Role:     m.Role,
			Nickname: m.Nickname,
			JoinTime: m.JoinTime,
			Online:   ws.ConnManager.IsOnline(user.ID),
		})
	}

	return result, nil
}

// JoinRoom 加入群聊
func JoinRoom(ctx context.Context, roomID, userID uint) error {
	// 检查群是否存在
	_, err := dao.GetRoomByID(ctx, roomID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return enum.CodeRoomNotFound
		}
		return err
	}

	// 检查是否已是群成员
	if dao.IsRoomMember(ctx, roomID, userID) {
		return nil // 已是成员，忽略
	}

	// 加入群
	err = dao.AddRoomMember(ctx, roomID, userID, int8(enum.RoomRoleMember))
	if err != nil {
		return err
	}

	// 更新成员数量
	_ = dao.UpdateRoomMemberCount(ctx, roomID)

	// 清除群缓存
	_ = cache.DelRoomMembersCache(ctx, roomID)

	return nil
}

// QuitRoom 退出群聊
func QuitRoom(ctx context.Context, roomID, userID uint) error {
	room, err := dao.GetRoomByID(ctx, roomID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return enum.CodeRoomNotFound
		}
		return err
	}

	// 群主不能退出群（只能解散群）
	if room.OwnerID == userID {
		return enum.CodeRoomOwnerCannotQuit
	}

	// 移除成员
	err = dao.RemoveRoomMember(ctx, roomID, userID)
	if err != nil {
		return err
	}

	// 更新成员数量
	_ = dao.UpdateRoomMemberCount(ctx, roomID)

	// 清除缓存
	_ = cache.DelRoomMembersCache(ctx, roomID)

	return nil
}

// DismissRoom 解散群聊（仅群主）
func DismissRoom(ctx context.Context, roomID, userID uint) error {
	room, err := dao.GetRoomByID(ctx, roomID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return enum.CodeRoomNotFound
		}
		return err
	}

	// 只有群主能解散
	if room.OwnerID != userID {
		return enum.CodeNoPermission
	}

	// 删除所有成员
	_ = dao.DeleteRoomAllMembers(ctx, roomID)

	// 删除群
	err = dao.DeleteRoom(ctx, roomID)
	if err != nil {
		return err
	}

	// 清除缓存
	_ = cache.DelRoomMembersCache(ctx, roomID)

	return nil
}

// UpdateRoom 更新群信息
func UpdateRoom(ctx context.Context, roomID uint, name, avatar, intro string, userID uint) error {
	room, err := dao.GetRoomByID(ctx, roomID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return enum.CodeRoomNotFound
		}
		return err
	}

	// 只有群主和管理员能更新
	if room.OwnerID != userID {
		return enum.CodeNoPermission
	}

	if name != "" {
		room.Name = name
	}
	if avatar != "" {
		room.Avatar = avatar
	}
	room.Intro = intro

	err = dao.UpdateRoom(ctx, room)
	if err != nil {
		return err
	}

	// 清除缓存
	_ = cache.DelRoomMembersCache(ctx, roomID)

	return nil
}

// SetRoomAdmin 设置/取消管理员
func SetRoomAdmin(ctx context.Context, roomID, targetUserID, operatorID uint, isAdmin bool) error {
	room, err := dao.GetRoomByID(ctx, roomID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return enum.CodeRoomNotFound
		}
		return err
	}

	// 只有群主能设置管理员
	if room.OwnerID != operatorID {
		return enum.CodeNoPermission
	}

	member, err := dao.GetRoomMember(ctx, roomID, targetUserID)
	if err != nil {
		return enum.CodeRoomMemberNotFound
	}

	if isAdmin {
		member.Role = int8(enum.RoomRoleAdmin)
	} else {
		member.Role = int8(enum.RoomRoleMember)
	}

	err = dao.UpdateRoomMember(ctx, member)
	if err != nil {
		return err
	}

	// 清除缓存
	_ = cache.DelRoomMembersCache(ctx, roomID)

	return nil
}

// KickRoomMember 踢出群成员
func KickRoomMember(ctx context.Context, roomID, targetUserID, operatorID uint) error {
	room, err := dao.GetRoomByID(ctx, roomID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return enum.CodeRoomNotFound
		}
		return err
	}

	// 群主或管理员能踢人
	isOwner := room.OwnerID == operatorID
	member, _ := dao.GetRoomMember(ctx, roomID, operatorID)
	isAdmin := member != nil && member.Role == int8(enum.RoomRoleAdmin)

	if !isOwner && !isAdmin {
		return enum.CodeNoPermission
	}

	// 不能踢群主
	if room.OwnerID == targetUserID {
		return enum.CodeNoPermission
	}

	// 不能踢管理员（除非是群主）
	targetMember, _ := dao.GetRoomMember(ctx, roomID, targetUserID)
	if targetMember != nil && targetMember.Role == int8(enum.RoomRoleAdmin) && !isOwner {
		return enum.CodeNoPermission
	}

	err = dao.RemoveRoomMember(ctx, roomID, targetUserID)
	if err != nil {
		return err
	}

	// 更新成员数量
	_ = dao.UpdateRoomMemberCount(ctx, roomID)

	// 清除缓存
	_ = cache.DelRoomMembersCache(ctx, roomID)

	return nil
}

// SearchRooms 搜索群
func SearchRooms(ctx context.Context, keyword string) ([]RoomInfo, error) {
	rooms, err := dao.SearchRooms(ctx, keyword)
	if err != nil {
		return nil, err
	}

	result := make([]RoomInfo, 0, len(rooms))
	for _, room := range rooms {
		avatar := room.Avatar
		if avatar == "" {
			avatar = fmt.Sprintf("/static/room/%d.png", room.ID)
		}
		result = append(result, RoomInfo{
			ID:         room.ID,
			Name:       room.Name,
			Avatar:     avatar,
			Intro:      room.Intro,
			OwnerID:    room.OwnerID,
			MemberCnt:  room.MemberCnt,
			CreatedAt:  room.CreatedAt.Unix(),
		})
	}

	return result, nil
}

// TransferRoomOwner 转让群主
func TransferRoomOwner(ctx context.Context, roomID, targetUserID, operatorID uint) error {
	room, err := dao.GetRoomByID(ctx, roomID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return enum.CodeRoomNotFound
		}
		return err
	}

	// 只有群主能转让
	if room.OwnerID != operatorID {
		return enum.CodeNoPermission
	}

	// 目标用户必须是群成员
	if !dao.IsRoomMember(ctx, roomID, targetUserID) {
		return enum.CodeRoomMemberNotFound
	}

	// 更新群主
	room.OwnerID = targetUserID
	err = dao.UpdateRoom(ctx, room)
	if err != nil {
		return err
	}

	// 更新转让者的角色为管理员
	member, _ := dao.GetRoomMember(ctx, roomID, operatorID)
	if member != nil {
		member.Role = int8(enum.RoomRoleAdmin)
		_ = dao.UpdateRoomMember(ctx, member)
	}

	// 清除缓存
	_ = cache.DelRoomMembersCache(ctx, roomID)

	return nil
}

// SaveGroupMessage 保存群消息到数据库
func SaveGroupMessage(ctx context.Context, roomID, fromID uint, content string, msgType int) error {
	lock := getRoomLock(roomID)
	lock.Lock()
	defer lock.Unlock()

	// 保存到消息表
	msg := &models.Message{
		FromID:      fromID,
		RoomID:      roomID,
		MessageType: int8(msgType),
		Content:     content,
	}
	return global.DB.WithContext(ctx).Create(msg).Error
}

// InviteMembers 邀请成员入群（仅群主和管理员）
func InviteMembers(ctx context.Context, roomID uint, targetUserIDs []uint, operatorID uint) (int, error) {
	room, err := dao.GetRoomByID(ctx, roomID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, enum.CodeRoomNotFound
		}
		return 0, err
	}

	// 检查权限：群主或管理员才能邀请
	isOwner := room.OwnerID == operatorID
	member, _ := dao.GetRoomMember(ctx, roomID, operatorID)
	isAdmin := member != nil && member.Role == int8(enum.RoomRoleAdmin)

	if !isOwner && !isAdmin {
		return 0, enum.CodeNoPermission
	}

	// 获取群信息用于通知
	roomAvatar := room.Avatar
	if roomAvatar == "" {
		roomAvatar = fmt.Sprintf("/static/room/%d.png", room.ID)
	}

	addedCount := 0
	for _, targetUserID := range targetUserIDs {
		// 检查目标用户是否已是群成员
		if dao.IsRoomMember(ctx, roomID, targetUserID) {
			continue
		}

		// 添加成员
		err = dao.AddRoomMember(ctx, roomID, targetUserID, int8(enum.RoomRoleMember))
		if err != nil {
			continue
		}
		addedCount++

		// 发送入群通知给被邀请者
		notifyData := map[string]interface{}{
			"msg_type": "room_invite",
			"data": map[string]interface{}{
				"room_id":     roomID,
				"room_name":   room.Name,
				"room_avatar": roomAvatar,
				"inviter_id":  operatorID,
			},
		}

		// 如果在线，发送WebSocket通知
		if ws.ConnManager.IsOnline(targetUserID) {
			client, _ := ws.ConnManager.Get(targetUserID)
			if client != nil {
				client.Wmu.Lock()
				_ = client.Conn.WriteJSON(notifyData)
				client.Wmu.Unlock()
			}
		} else {
			// 离线消息
			_ = cache.SaveOfflineMessage(ctx, targetUserID, notifyData)
		}
	}

	// 更新成员数量
	if addedCount > 0 {
		_ = dao.UpdateRoomMemberCount(ctx, roomID)
		_ = cache.DelRoomMembersCache(ctx, roomID)
	}

	return addedCount, nil
}

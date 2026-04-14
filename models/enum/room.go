package enum

type RoomRole int8

const (
	RoomRoleOwner  RoomRole = 0 // 群主
	RoomRoleAdmin  RoomRole = 1 // 管理员
	RoomRoleMember RoomRole = 2 // 普通成员
)

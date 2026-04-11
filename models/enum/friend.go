package enum

type FriendStatus int8

const (
	FriendNormal  FriendStatus = 1 // 正常
	FriendBlack   FriendStatus = 2 // 拉黑
	FriendDeleted FriendStatus = 3 // 删除
)

package request

type FriendRequest struct {
	FriendID uint `json:"friend_id"`
}
type FriendRequestOK struct {
	FromID uint `json:"from_id"`
}

// RejectFriendRequest 拒绝好友请求
type RejectFriendRequest struct {
	FromID uint `json:"from_id"`
}

// DeleteFriend 删除好友
type DeleteFriendRequest struct {
	FriendID uint `json:"friend_id"`
}

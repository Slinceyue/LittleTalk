package request

type FriendRequest struct {
	FriendID uint `json:"friend_id"`
}
type FriendRequestOK struct {
	FromID uint `json:"from_id"`
}

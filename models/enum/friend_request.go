package enum

type FriendRequestStatus int8

const (
	FriendRequestStatusPending  FriendRequestStatus = 0  // 待处理（未读/未同意/未拒绝，默认）
	FriendRequestStatusAccepted FriendRequestStatus = 1  // 已同意（添加好友成功）
	FriendRequestStatusRejected FriendRequestStatus = -1 // 已拒绝
)

// String 实现Stringer接口，方便日志/返回中文描述
func (s FriendRequestStatus) String() string {
	switch s {
	case FriendRequestStatusPending:
		return "待处理"
	case FriendRequestStatusAccepted:
		return "已同意"
	case FriendRequestStatusRejected:
		return "已拒绝"
	default:
		return "未知状态"
	}
}

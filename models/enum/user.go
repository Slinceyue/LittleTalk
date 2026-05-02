package enum

type Sex int8

const (
	Unknown    Sex = 0
	MaleType   Sex = 1
	FemaleType Sex = 2
)

func (s Sex) int() int8 {
	return int8(s)
}
func (s Sex) String() string {
	switch s {
	case MaleType:
		return "男"
	case FemaleType:
		return "女"
	default:
		return "未知"
	}
}

type Role int8

const (
	// 普通用户
	UserRole Role = 1
	// 管理员
	AdminRole Role = 2
)

func (r Role) String() string {
	switch r {
	case UserRole:
		return "用户"
	case AdminRole:
		return "管理员"
	default:
		return "用户"
	}
}

type OnlineStatus int8

const (
	Offline OnlineStatus = 0 // 离线
	Online  OnlineStatus = 1 // 在线
)

func (s OnlineStatus) String() string {
	switch s {
	case Online:
		return "在线"
	case Offline:
		return "离线"
	default:
		return "未知"
	}
}

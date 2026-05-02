package request

type NewUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	BackSalt string `json:"back_salt"`
	Sex      int8   `json:"sex"`
	Birthday string `json:"birthday"`
}
type LoginRequest struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Password string `json:"password"`
}
type UserUpdateRequest struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Sex      int8   `json:"sex"`
	Intro    string `json:"intro"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Birthday string `json:"birthday"`
}

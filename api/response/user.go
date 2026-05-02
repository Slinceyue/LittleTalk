package response

type SelfUserResponse struct {
	ID       uint   `json:"id"`
	Avatar   string `json:"avatar"`
	Username string `json:"username"`
	Sex      int8   `json:"sex"`
	Intro    string `json:"intro"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Birthday string `json:"birthday"`
	Status   int8   `json:"status"`
}

type OtherUserResponse struct {
	ID       uint   `json:"id"`
	Avatar   string `json:"avatar"`
	Username string `json:"username"`
	Sex      int8   `json:"sex"`
	Intro    string `json:"intro"`
	Birthday string `json:"birthday"`
}

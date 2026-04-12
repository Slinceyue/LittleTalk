package resbonse

import "LittleTalk/models/enum"

type Response struct {
	Code    enum.ResCode `json:"code"`
	Message string       `json:"message"`
	Data    interface{}  `json:"data"`
}

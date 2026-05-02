package message_handler

import (
	"LittleTalk/api/request"
	"LittleTalk/cache"
	"LittleTalk/service"
	"LittleTalk/utils/ws"
	"encoding/json"

	"github.com/gin-gonic/gin"
)

type MessageHandler struct{}

func (MessageHandler) WS(c *gin.Context) {
	userID, _ := c.Get("id")
	ctx := c.Request.Context()
	conn, err := ws.Upgrader.Upgrade(c.Writer, c.Request, nil)
	ws.ConnManager.Add(userID.(uint), conn)
	if err != nil {
		return
	}
	defer conn.Close()
	defer cache.SetUserOfflineStatus(ctx, userID.(uint))
	_ = cache.SetUserOnlineStatus(c, userID.(uint))
	go func() {
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				break
			}
			var message request.TalkMessageRequest
			json.Unmarshal(data, &message)
			service.MessageChannel <- &request.MessageContext{
				MsgType:            "talk",
				TalkMessageRequest: message,
				Ctx:                ctx,
			}
		}
	}()
}

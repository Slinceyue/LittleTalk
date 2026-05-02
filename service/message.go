package service

import (
	"LittleTalk/api/request"
	"LittleTalk/cache"
	"LittleTalk/utils/ws"
	"context"
	"errors"

	"github.com/sirupsen/logrus"
)

var MessageChannel = make(chan *request.MessageContext, 1000)
var senderCtx, senderCancel = context.WithCancel(context.Background())

func Run() {
	go MessageSender()
}
func MessageSender() {
	for {
		select {
		case <-senderCtx.Done():
			return
		case msg, ok := <-MessageChannel:
			if !ok {
				return
			}

			err := Send(*msg)
			if err != nil {
				logrus.Error(err.Error())
				continue
			}
		}
	}
}

// 停服时调用
func StopMessageSender() {
	senderCancel()
	close(MessageChannel)
}
func IsUserOnline(userID uint) bool {
	if !cache.GetUserOnlineStatus(context.Background(), userID) {
		return false
	}
	_, ok := ws.ConnManager.Get(userID)
	if !ok {
		return false
	}
	return true
}
func Send(msg request.MessageContext) error {
	var toID uint
	var msgSend interface{}
	switch msg.MsgType {
	case "friend":
		{
			toID = msg.FriendMessageRequest.ToID
			msgSend = msg.FriendMessageRequest
		}
	case "talk":
		{
			toID = msg.TalkMessageRequest.ToID
			msgSend = msg.TalkMessageRequest
		}
	}
	if toID == 0 {
		return errors.New("用户不存在")
	}
	client, _ := ws.ConnManager.Get(toID)
	if client == nil {
		return errors.New("用户不在线")
	}
	// 2. 连接级写锁（串行化写入）
	client.Wmu.Lock()
	// 3. 带超时写入（核心优化）
	sender := map[string]interface{}{
		"msg_type": msg.MsgType,
		"data":     msgSend,
	}
	err := client.Conn.WriteJSON(sender)
	if err != nil {
		_ = client.Conn.Close()
		ws.ConnManager.Delete(toID)
		return errors.New("连接已断开")
	}
	client.Wmu.Unlock() // 写入完成立即解锁
	return err
}

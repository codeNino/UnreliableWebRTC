package handler

import (
	"time"
)

func StartSendingBinaryMessages() {
	ticker := time.NewTicker(time.Duration(int64(time.Millisecond) * 17))
	defer ticker.Stop()

	for range ticker.C {
		SendBinaryMessageToAllClients("Hello")
	}
}

func SendBinaryMessageToAllClients(message string) {
	Players.Range(func(key, value interface{}) bool {
		p := value.(*player)
		p.datachan.Send([]byte(message))
		return true
	})
}

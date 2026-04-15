package realtime

import "fmt"

func PlayRoomTopic(roomCode string) string {
	return fmt.Sprintf("stream:play:room:%s", roomCode)
}

func PlayRoomBetsTopic(roomCode string, userID int64) string {
	return fmt.Sprintf("stream:play:bets:%s:user:%d", roomCode, userID)
}

func WalletUserTopic(userID int64) string {
	return fmt.Sprintf("stream:wallet:user:%d", userID)
}

func AdminRoomsTopic() string {
	return "stream:admin:rooms"
}

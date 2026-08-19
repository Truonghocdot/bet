package realtime

import "fmt"

func UserEventTopic(userID int64) string {
	return fmt.Sprintf("stream:user:%d", userID)
}

func WheelSessionTopic(sessionID int64) string {
	return fmt.Sprintf("stream:wheel:session:%d", sessionID)
}

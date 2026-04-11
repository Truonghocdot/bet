package clock

import (
	"sync"
	"time"
)

const vietnamLocationName = "Asia/Ho_Chi_Minh"

var (
	locationOnce sync.Once
	location     *time.Location
)

func Location() *time.Location {
	locationOnce.Do(func() {
		loc, err := time.LoadLocation(vietnamLocationName)
		if err != nil {
			location = time.FixedZone(vietnamLocationName, 7*60*60)
			return
		}
		location = loc
	})
	return location
}

func Now() time.Time {
	return time.Now().In(Location())
}


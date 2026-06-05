package room

import "time"

type Room struct {
	Number       int
	Name         string
	IsActive     bool
	WeekdayOpen  int
	WeekdayClose int
	FriSatOpen   int
	FriSatClose  int
	SunOpen      int
	SunClose     int
	CreatedAt    time.Time
}

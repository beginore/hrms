package clock

import "time"

type Clock struct{}

func NewClock() *Clock {
	return &Clock{}
}

func (clock *Clock) Now() time.Time {
	return time.Now()
}

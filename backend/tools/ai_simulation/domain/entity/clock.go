package entity

import "time"

// SimClock tracks simulated time as tick count, where 1 tick = 1 minute
// (per the sprint brief). Kept separate from wall-clock time entirely — a
// 30-day run executes as fast as the host machine allows, not in real time.
type SimClock struct {
	StartDate time.Time // simulated calendar start (for weekday/holiday logic)
	Tick      int64     // minutes elapsed since StartDate
}

func NewSimClock(start time.Time) *SimClock {
	return &SimClock{StartDate: start, Tick: 0}
}

// Now returns the current simulated timestamp.
func (c *SimClock) Now() time.Time {
	return c.StartDate.Add(time.Duration(c.Tick) * time.Minute)
}

// Advance moves the clock forward by one tick (one simulated minute).
func (c *SimClock) Advance() {
	c.Tick++
}

// Day returns the 0-indexed simulated day number.
func (c *SimClock) Day() int {
	return int(c.Tick / (24 * 60))
}

// MinuteOfDay returns 0-1439, the minute within the current simulated day.
func (c *SimClock) MinuteOfDay() int {
	return int(c.Tick % (24 * 60))
}

// Hour returns 0-23.
func (c *SimClock) Hour() int {
	return c.MinuteOfDay() / 60
}

// IsWeekend reports whether the current simulated date is Saturday or Sunday.
func (c *SimClock) IsWeekend() bool {
	wd := c.Now().Weekday()
	return wd == time.Saturday || wd == time.Sunday
}

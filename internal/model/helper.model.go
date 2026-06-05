package model

import (
	"time"
)

// CompactDate fits Year, Month, and Day into a 4-byte uint32
type CompactDate uint32

// NewCompactDate packs a date into 4 bytes
func NewCompactDate(year int, month int, day int) CompactDate {
	return CompactDate((year << 9) | (month << 5) | day)
}

// Unpack extracts the year, month, and day using bitwise masks
func (cd CompactDate) Unpack() (year int, month int, day int) {
	// Day is in the lowest 5 bits (mask 0x1F or 31)
	day = int(cd & 0x1F)

	// Month is in the next 4 bits, shift right by 5 to isolate it (mask 0x0F or 15)
	month = int((cd >> 5) & 0x0F)

	// Year is in the remaining upper bits, shift right by 9
	year = int(cd >> 9)

	return year, month, day
}

// AsTime converts the compact date back into a standard Go time.Time object
func (cd CompactDate) AsTime() time.Time {
	y, m, d := cd.Unpack()
	return time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
}

// DaysBetween returns how many days have elapsed from 'from' to 'to'.
// Returns a negative value if 'to' is before 'from'.
func DaysBetween(from, to CompactDate) int {
	return int(to.AsTime().Sub(from.AsTime()).Hours() / 24)
}

// AddDays returns a new CompactDate advanced by the given number of days.
func (cd CompactDate) AddDays(days int) CompactDate {
	t := cd.AsTime().AddDate(0, 0, days)
	return NewCompactDate(t.Year(), int(t.Month()), t.Day())
}

// ExpiryDate computes the expiry date given a purchase date, base expiry days, and a storage multiplier.
// A multiplier <= 0 is treated as 1 (room temperature, no extension).
func ExpiryDate(purchaseDate CompactDate, baseExpiryDays uint16, multiplier float32) CompactDate {
	m := multiplier
	if m <= 0 {
		m = 1
	}
	totalDays := int(float32(baseExpiryDays) * m)
	t := purchaseDate.AsTime().AddDate(0, 0, totalDays)
	return NewCompactDate(t.Year(), int(t.Month()), t.Day())
}

// IsExpired returns true if the food has passed its expiry date given the current date.
func IsExpired(purchaseDate CompactDate, baseExpiryDays uint16, multiplier float32, currentDate CompactDate) bool {
	expiry := ExpiryDate(purchaseDate, baseExpiryDays, multiplier)
	return DaysBetween(expiry, currentDate) > 0
}

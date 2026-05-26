package model

import (
	"fmt"
	"time"
	"unsafe"
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

func main() {
	// 1. Pack "01/02/15" (Assuming Feb 1st, 2015)
	packed := NewCompactDate(2015, 2, 1)
	fmt.Printf("Packed Size: %d bytes\n", unsafe.Sizeof(packed)) // 4 bytes

	// 2. Unpack raw integers
	y, m, d := packed.Unpack()
	fmt.Printf("Unpacked: Year=%d, Month=%d, Day=%d\n", y, m, d)

	// 3. Convert to time.Time for standard Go date operations
	t := packed.AsTime()
	fmt.Println("As time.Time:", t.Format("2006-01-02"))
}

package core

import (
	"fmt"

	"github.com/kamil5b/basic-life-sim/internal/model"
)

// advanceTime moves the game clock forward by the given number of minutes.
func advanceTime(char *model.Character, minutes float64) {
	if minutes <= 0 {
		return
	}
	char.SubMinute += minutes
	wholeMinutes := int(char.SubMinute + 1e-9)
	if wholeMinutes == 0 {
		return
	}
	char.SubMinute -= float64(wholeMinutes)

	totalMinutes := float64(char.Hour)*60 + float64(char.Minute) + float64(wholeMinutes)
	hours := int(totalMinutes / 60)
	mins := int(totalMinutes) % 60
	char.Hour = uint8(hours % 24)
	char.Minute = uint8(mins)
	daysPassed := hours / 24
	if daysPassed > 0 {
		char.CurrentDate = char.CurrentDate.AddDays(daysPassed)
	}
}

// timeOfDay returns whether it is currently Day or Night.
func timeOfDay(char *model.Character) string {
	if char.Hour >= 6 && char.Hour < 18 {
		return "Day"
	}
	return "Night"
}

// formatGameTime returns a human-readable date/time string.
func formatGameTime(char *model.Character) string {
	y, m, d := char.CurrentDate.Unpack()
	return fmt.Sprintf("%04d-%02d-%02d %02d:%02d (%s)", y, m, d, char.Hour, char.Minute, timeOfDay(char))
}

// canSkipTime reports whether an action supports a time-skip prompt.
func canSkipTime(action string) bool {
	switch action {
	case "sleep", "watch", "stream":
		return true
	}
	return false
}

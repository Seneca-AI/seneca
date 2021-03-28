package util

import (
	"fmt"
	"seneca/api/senecaerror"
	"strings"
	"time"
)

// GetFileNameFromPath parses the path and extracts the last string after '/'.
// Params:
// 		path string: the path
// Returns:
//		error if path is an empty string
func GetFileNameFromPath(path string) (string, error) {
	if path == "" {
		return "", senecaerror.NewBadStateError(fmt.Errorf("received empty string"))
	}
	pathSplit := strings.Split(path, "/")
	return pathSplit[len(pathSplit)-1], nil
}

// TimeToMilliseconds gets the unix time in milliseconds from the give time.Time.
// Params:
//		t *time.Time
// Returns:
//		int64: milliseconds
func TimeToMilliseconds(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}

// DurationToMilliseconds converts the time.Duration to milliseconds.
// Params:
//		t *time.Duration
// Returns:
//		int64: milliseconds
func DurationToMilliseconds(t time.Duration) int64 {
	return t.Milliseconds() / int64(time.Millisecond)
}

// MillisecondsToTime converts milliseconds to a time.Time object.
// Params:
//		ms int64
// Returns:
//		time.Time
func MillisecondsToTime(ms int64) time.Time {
	return time.Unix(0, ms*int64(time.Millisecond))
}

// MillisecondsToDuration converts milliseconds to a time.Duration object.
// Params:
//		ms int64
// Returns:
//		time.Duration
func MillisecondsToDuration(ms int64) time.Duration {
	return time.Unix(0, ms*int64(time.Millisecond)).Sub(time.Unix(0, 0))
}

// DurationToString returns the time.Duration in the string form hh:mm:ss.
// Params:
//		dur time.Duration
// Returns:
//		string
func DurationToString(dur time.Duration) string {
	rounded := dur.Round(time.Second)
	hour := rounded / time.Hour
	hourString := fmt.Sprintf("%d", hour)
	if hour < 10 {
		hourString = "0" + hourString
	}

	rounded -= hour * time.Hour
	minute := rounded / time.Minute
	minuteString := fmt.Sprintf("%d", minute)
	if minute < 10 {
		minuteString = "0" + minuteString
	}

	rounded -= minute * time.Minute
	second := rounded / time.Second
	secondString := fmt.Sprintf("%d", second)
	if second < 10 {
		secondString = "0" + secondString
	}

	return fmt.Sprintf("%s:%s:%s", hourString, minuteString, secondString)
}

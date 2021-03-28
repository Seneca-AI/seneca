package util

import (
	"fmt"
	"os"
	"seneca/api/senecaerror"
	"seneca/api/types"
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

// LocationsEquals compares the degrees and direction of the locations.
// Params:
//		l1 *types.Location
//		l2 *types.Location
// Returns:
//		bool
func LocationsEquals(l1 *types.Location, l2 *types.Location) bool {
	if l1 == nil || l2 == nil {
		return l1 == l2
	}
	if l1.Lat == nil || l2.Lat == nil {
		return l1.Lat == l2.Lat
	}
	if l1.Long == nil || l2.Long == nil {
		return l1.Long == l2.Long
	}
	return l1.Lat.Degrees == l2.Lat.Degrees && l1.Lat.DegreeMinutes == l2.Lat.DegreeMinutes && l1.Lat.DegreeSeconds == l2.Lat.DegreeSeconds && l1.Lat.LatDirection == l2.Lat.LatDirection &&
		l1.Long.Degrees == l2.Long.Degrees && l1.Long.DegreeMinutes == l2.Long.DegreeMinutes && l1.Long.DegreeSeconds == l2.Long.DegreeSeconds && l1.Long.LongDirection == l2.Long.LongDirection
}

// IsCIEnv returns true if the env variable "CI" is set to "true".
func IsCIEnv() bool {
	val, ok := os.LookupEnv("CI")
	if ok && val == "true" {
		return true
	}
	return false
}

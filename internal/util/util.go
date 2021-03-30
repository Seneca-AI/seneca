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

// LocationsEqual compares the degrees and direction of the locations.
// Params:
//		l1 *types.Location
//		l2 *types.Location
// Returns:
//		bool
func LocationsEqual(l1 *types.Location, l2 *types.Location) bool {
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

func MotionsEqual(m1 *types.Motion, m2 *types.Motion) bool {
	if m1 == nil || m2 == nil {
		return m1 == m2
	}
	return m1.VelocityMph == m2.VelocityMph && m1.AccelerationMphS == m2.AccelerationMphS
}

// IsCIEnv returns true if the env variable "CI" is set to "true".
func IsCIEnv() bool {
	val, ok := os.LookupEnv("CI")
	if ok && val == "true" {
		return true
	}
	return false
}

//	ConstructRawLocationDatas construct a list of types.RawLocation from a list of types.Location and time.Time for the given userID.
//	Params:
//		userID string
//		locations []*types.Location
//		times	[]time.Time
//	Returns:
//		[]*types.RawLocation
//		error
func ConstructRawLocationDatas(userID string, locations []*types.Location, times []time.Time) ([]*types.RawLocation, error) {
	if len(locations) != len(times) {
		return nil, fmt.Errorf("locations has length %d, but times has legth %d", len(locations), len(times))
	}
	rawLocations := []*types.RawLocation{}
	for i := range locations {
		rawLocations = append(rawLocations, &types.RawLocation{
			UserId:      userID,
			Location:    locations[i],
			TimestampMs: TimeToMilliseconds(times[i]),
		})
	}
	return rawLocations, nil
}

//	ConstructRawMotionDatas construct a list of types.RawMotion from a list of types.Motion and time.Time for the given userID.
//	Params:
//		userID string
//		motions []*types.Motion
//		times	[]time.Time
//	Returns:
//		[]*types.RawMotion
//		error
func ConstructRawMotionDatas(userID string, motions []*types.Motion, times []time.Time) ([]*types.RawMotion, error) {
	if len(motions) != len(times) {
		return nil, fmt.Errorf("motions has length %d, but times has legth %d", len(motions), len(times))
	}
	rawMotions := []*types.RawMotion{}
	for i := range motions {
		rawMotions = append(rawMotions, &types.RawMotion{
			UserId:      userID,
			Motion:      motions[i],
			TimestampMs: TimeToMilliseconds(times[i]),
		})
	}
	return rawMotions, nil
}

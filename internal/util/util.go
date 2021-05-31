package util

import (
	"fmt"
	"math/rand"
	"os"
	"seneca/api/senecaerror"
	st "seneca/api/type"
	"strconv"
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

// MillisecondsToTime converts milliseconds to a time.Time object.
// Params:
//		ms int64
// Returns:
//		time.Time
func MillisecondsToTime(ms int64) time.Time {
	return time.Unix(0, ms*int64(time.Millisecond)).In(time.UTC)
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

// IsCIEnv returns true if the env variable "CI" is set to "true".
func IsCIEnv() bool {
	val, ok := os.LookupEnv("CI")
	if ok && val == "true" {
		return true
	}
	return false
}

func GenerateRandID() string {
	id := rand.Int63()
	return strconv.FormatInt(id, 10)
}

func EventExternalToPrettyString(event *st.Event) string {
	output := "Event: {\n"
	output += fmt.Sprintf("  EventType: %s,\n", event.EventType)
	output += fmt.Sprintf("  Value: %f,\n", event.Value)
	output += fmt.Sprintf("  Severity: %f,\n", event.Severity)
	output += fmt.Sprintf("  Timestamp: %v,\n", MillisecondsToTime(event.TimestampMs))
	output += "  Source: {\n"
	output += fmt.Sprintf("    SourceType: %s,\n", event.ExternalSource.SourceType)
	output += fmt.Sprintf("    VideoURL: %s,\n", event.ExternalSource.VideoUrl)
	output += "  }\n"
	output += "}"

	return output
}

func DrivingConditionExternalToPrettyString(dc *st.DrivingCondition) string {
	output := "DrivingCondition: {\n"
	output += fmt.Sprintf("  TimePeriod: [%v - %v],\n", MillisecondsToTime(dc.StartTimeMs), MillisecondsToTime(dc.EndTimeMs))
	output += fmt.Sprintf("  Conditions: %12s,\n", dc.ConditionType)
	output += fmt.Sprintf("  Severities: %12f,\n", dc.Severity)
	output += "  Source: [\n"
	for _, eSrc := range dc.ExternalSource {
		output += fmt.Sprintf("    (%s, %s),\n", eSrc.SourceType, eSrc.VideoUrl)
	}
	output += "  ],\n"
	output += "}"

	return output
}

func StringToInterfaceMapOrFalse(key string, mapObj map[string]interface{}) (map[string]interface{}, bool) {
	subMapObj, ok := mapObj[key]
	if !ok {
		return nil, false
	}
	subMap, ok := subMapObj.(map[string]interface{})
	if !ok {
		return nil, false
	}
	return subMap, true
}

func RemoveTrailingZeroes(in string) string {
	for i := len(in) - 1; i > 0; i-- {
		if in[i:i+1] == "0" || in[i:i+1] == "." {
			in = in[:i]
		} else {
			break
		}
	}
	return in
}

package headerparse

import (
	"fmt"
	"math"
	"seneca/api/constants"
	st "seneca/api/type"
	"sort"
	"strconv"
	"strings"
	"time"
)

// In the form "<deg> deg <deg_mins>' <deg_sconds>\"".
func parseDegrees(degreesStr string) (int32, int32, float64, error) {
	degreesStrSplit := strings.Split(degreesStr, " ")
	if len(degreesStrSplit) != 4 {
		return 0, 0, 0, fmt.Errorf("invalid length for lat/long degrees string %q", degreesStr)
	}

	if degreesStrSplit[1] != "deg" {
		return 0, 0, 0, fmt.Errorf("invalid format for latitude string %q", degreesStr)
	}
	degrees64, err := strconv.ParseInt(degreesStrSplit[0], 10, 32)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("error parsing float from deg in %q", degreesStr)
	}
	degrees := int32(degrees64)

	if degreesStrSplit[2][len(degreesStrSplit[2])-1:] != "'" {
		return 0, 0, 0, fmt.Errorf("error parsing lat/long degrees string %q, missing degreeMins symbol", degreesStr)
	}
	degreeMins64, err := strconv.ParseInt(degreesStrSplit[2][:len(degreesStrSplit[2])-1], 10, 32)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("error parsing float from deg mins in %q", degreesStr)
	}
	degreeMins := int32(degreeMins64)

	if degreesStrSplit[3][len(degreesStrSplit[3])-1:] != "\"" {
		return 0, 0, 0, fmt.Errorf("error parsing lat/long degrees string %q, missing degreeSecs symbol - %q != %q", degreesStr, degreesStrSplit[3][len(degreesStrSplit[3])-1:], "\\\"")
	}
	degreeSecs, err := strconv.ParseFloat(degreesStrSplit[3][:len(degreesStrSplit[3])-1], 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("error parsing float from deg secs in %q", degreesStr)
	}

	return degrees, degreeMins, degreeSecs, nil
}

func getDurationMs(durationString string) (int64, error) {
	durationString = strings.Replace(durationString, ":", "h", 1)
	durationString = strings.Replace(durationString, ":", "m", 1)
	durationString = durationString + "s"
	duration, err := time.ParseDuration(durationString)
	if err != nil {
		return 0, fmt.Errorf("error parsing duration - err: %w", err)
	}
	if duration.Milliseconds() == 0 {
		return 0, fmt.Errorf("duration from durationString %q is zero", durationString)
	}
	return duration.Milliseconds(), nil
}

// Used for sorting by time.
type locationMotionTime struct {
	location *st.Location
	motion   *st.Motion
	gpsTime  time.Time
}

func getLocationMotionTime(gpsDateTimeFormat string, unprocessedGPSData *unprocessedExifGPSData) (*locationMotionTime, error) {
	var err error

	locationMotionTime := &locationMotionTime{
		location: &st.Location{},
		motion:   &st.Motion{},
		gpsTime:  time.Time{},
	}

	if locationMotionTime.location.Lat, err = stringToLatitude(unprocessedGPSData.latitude); err != nil {
		return nil, fmt.Errorf("error parsing Latitude: %w", err)
	}
	if locationMotionTime.location.Long, err = stringToLongitude(unprocessedGPSData.longitude); err != nil {
		return nil, fmt.Errorf("error parsing Longitude: %w", err)
	}

	if locationMotionTime.gpsTime, err = time.Parse(gpsDateTimeFormat, unprocessedGPSData.datetime); err == nil {
		locationMotionTime.gpsTime = locationMotionTime.gpsTime.In(time.UTC).Round(time.Second)
	}

	if err != nil {
		return nil, fmt.Errorf("error parsing GPS time: %w", err)
	}

	switch unprocessedGPSData.speedRef {
	case "mph":
		locationMotionTime.motion.VelocityMph = unprocessedGPSData.speed
	case "km/h":
		locationMotionTime.motion.VelocityMph = math.Round(unprocessedGPSData.speed / constants.KilometersToMiles)
	default:
		return nil, fmt.Errorf("invalid speedRef %q", unprocessedGPSData.speedRef)
	}

	return locationMotionTime, nil
}

// 	getLocationsMotionsTimes extracts a list of st.Location, st.Motion and time.Time from the video at the given path.
func getLocationsMotionsTimes(gpsDateTimeFormat string, unprocessedExifData *unprocessedExifData) ([]*st.Location, []*st.Motion, []time.Time, error) {
	locationsMotionsTimes := []locationMotionTime{}
	for _, gpsData := range unprocessedExifData.gpsData {
		lmt, err := getLocationMotionTime(gpsDateTimeFormat, gpsData)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("error extracting GPS data: %w", err)
		}
		locationsMotionsTimes = append(locationsMotionsTimes, *lmt)
	}

	sort.Slice(locationsMotionsTimes, func(i, j int) bool { return locationsMotionsTimes[i].gpsTime.Before(locationsMotionsTimes[j].gpsTime) })
	locations := []*st.Location{}
	motions := []*st.Motion{}
	times := []time.Time{}
	for _, lmt := range locationsMotionsTimes {
		locations = append(locations, lmt.location)
		motions = append(motions, lmt.motion)
		times = append(times, lmt.gpsTime)
	}
	populateAccelerations(motions)

	return locations, motions, times, nil
}

package headerparse

import (
	"errors"
	"fmt"
	"math"
	"seneca/api/constants"
	st "seneca/api/type"
	"seneca/internal/util"
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

func getCreationTime(dashCamName DashCamName, timeString string) (int64, error) {
	t, err := time.Parse(cameraDataLayouts[dashCamName].videoStartTimeLayout, timeString)
	if err != nil {
		return 0, fmt.Errorf("error parsing CreationTime - err: %v", err)
	}
	t = t.In(time.UTC).Round(time.Second)

	if t.Equal(time.Unix(0, 0)) {
		return 0, errors.New("creationTime of 0 is not allowed")
	}
	t = t.In(time.UTC)

	return util.TimeToMilliseconds(t), nil
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

func getLocationMotionTime(unprocessedGPSData *unprocessedExifGPSData) (*locationMotionTime, error) {
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

	for _, layout := range gpsDateTimeParseLayouts {
		if locationMotionTime.gpsTime, err = time.Parse(layout, unprocessedGPSData.datetime); err == nil {
			locationMotionTime.gpsTime = locationMotionTime.gpsTime.In(time.UTC).Round(time.Second)
			break
		}
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
		return nil, fmt.Errorf("invalid spedRef %q", unprocessedGPSData.speedRef)
	}

	return locationMotionTime, nil
}

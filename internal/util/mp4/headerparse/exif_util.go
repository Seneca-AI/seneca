package headerparse

import (
	"fmt"
	"seneca/api/senecaerror"
	st "seneca/api/type"
	"seneca/internal/util"
	"seneca/internal/util/data"
	"strings"
	"time"

	"gopkg.in/ugjka/go-tz.v2/tz"
)

func stringToLatitude(latString string) (*st.Latitude, error) {
	latStringSplit := strings.Split(latString, " ")
	if len(latStringSplit) != 5 {
		return nil, fmt.Errorf("invalid format for longitude string %q", latString)
	}

	degrees, degreeMins, degreeSecs, err := parseDegrees(strings.Join(latStringSplit[:4], " "))
	if err != nil {
		return nil, fmt.Errorf("error parsing degrees from latString - err: %w", err)
	}

	latitude := &st.Latitude{
		Degrees:       degrees,
		DegreeMinutes: degreeMins,
		DegreeSeconds: degreeSecs,
	}

	switch latStringSplit[4] {
	case "N":
		latitude.LatDirection = st.Latitude_NORTH
	case "S":
		latitude.LatDirection = st.Latitude_SOUTH
	default:
		return nil, fmt.Errorf("error parsing latitude direction from %q", latString)
	}

	return latitude, nil
}

func stringToLongitude(longString string) (*st.Longitude, error) {
	longStringSplit := strings.Split(longString, " ")
	if len(longStringSplit) != 5 {
		return nil, fmt.Errorf("invalid format for longitude string %q", longString)
	}

	degrees, degreeMins, degreeSecs, err := parseDegrees(strings.Join(longStringSplit[:4], " "))
	if err != nil {
		return nil, fmt.Errorf("error parsing degrees from longString - err: %w", err)
	}

	longitude := &st.Longitude{
		Degrees:       degrees,
		DegreeMinutes: degreeMins,
		DegreeSeconds: degreeSecs,
	}

	switch longStringSplit[4] {
	case "E":
		longitude.LongDirection = st.Longitude_EAST
	case "W":
		longitude.LongDirection = st.Longitude_WEST
	default:
		return nil, fmt.Errorf("error parsing longitude direction from %q", longString)
	}

	return longitude, nil
}

func validateData(rawVideo *st.RawVideo, locations []*st.Location, motions []*st.Motion, times []time.Time) error {
	if len(times) == 0 {
		return senecaerror.NewUserError("", fmt.Errorf("no GPS data found"), "Video has no header data.")
	}

	if times[0].Before(util.MillisecondsToTime(rawVideo.CreateTimeMs)) {
		return senecaerror.NewDevError(fmt.Errorf("location/motion data timestamp %v is before rawVideo createTime %v", times[0], util.MillisecondsToTime(rawVideo.CreateTimeMs).In(time.UTC)))
	}

	if times[len(times)-1].After(util.MillisecondsToTime(rawVideo.CreateTimeMs + rawVideo.DurationMs + time.Second.Milliseconds())) {
		return senecaerror.NewDevError(fmt.Errorf("location/motion data timestamp %v is after rawVideo endTime %v", times[0], util.MillisecondsToTime(rawVideo.CreateTimeMs+rawVideo.DurationMs).In(time.UTC)))
	}

	return nil
}

func populateAccelerations(motions []*st.Motion) {
	if len(motions) == 0 {
		return
	}

	motions[0].AccelerationMphS = 0

	for i := 1; i < len(motions); i++ {
		motions[i].AccelerationMphS = motions[i].VelocityMph - motions[i-1].VelocityMph
	}
}

func stringExists(key string, data map[string]interface{}) (string, bool) {
	valObj, ok := data[key]
	if !ok {
		return "", false
	}
	val, ok := valObj.(string)
	return val, ok
}

func getTZOffset(t time.Time, location *st.Location) (time.Duration, error) {
	latFloat64 := data.LatitudeToFloat64(location.Lat)
	longFloat64 := data.LongitudeToFloat64(location.Long)

	timeZoneIDs, err := tz.GetZone(tz.Point{Lat: latFloat64, Lon: longFloat64})
	if err != nil {
		return 0, fmt.Errorf("tz.GetZone(%f, %f) returns err: %w", latFloat64, longFloat64, err)
	}

	if len(timeZoneIDs) == 0 {
		return 0, fmt.Errorf("tz.GetZone(%f, %f) returns 0 timeZoneIDs", latFloat64, longFloat64)
	}

	tzLocation, err := time.LoadLocation(timeZoneIDs[0])
	if err != nil {
		return 0, fmt.Errorf("time.LoadLocation(%s) returns err: %w", timeZoneIDs[0], err)
	}

	_, offset := t.In(tzLocation).Zone()

	return time.Second * time.Duration(offset), nil
}

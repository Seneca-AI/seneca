package headerparse

import (
	"fmt"
	"seneca/api/senecaerror"
	st "seneca/api/type"
	"seneca/internal/util"
	"strings"
	"time"
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

	fmt.Printf("DEBUG video create tiem: %v\n", util.MillisecondsToTime(rawVideo.CreateTimeMs))
	fmt.Printf("DEBUG first time: %v\n", times[0])
	fmt.Printf("DEBUG last time: %v\n", times[len(times)-1])

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

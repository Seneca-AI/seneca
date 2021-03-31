package mp4

import (
	"fmt"
	"seneca/api/types"
	"strconv"
	"strings"
	"time"
)

// MP4ToolInterface defines the interface for interacting with MP4 files
// throughout Seneca.
//nolint
type MP4ToolInterface interface {
	// ParseOutRawVideoMetadata extracts *types.RawVideo metadata from the mp4 at the given path.
	// Params:
	// 		string pathToVideo: path to mp4 to get RawVideo metadata from
	// Returns:
	// 		*types.RawVideo: the RawVideo object
	//		error
	ParseOutRawVideoMetadata(pathToVideo string) (*types.RawVideo, error)
	// 	ParseOutGPSMetadata extracts a list of types.Location, types.Motion and time.Time from the video at the given path.
	//	Params:
	//		 pathToVideo string: the video to analyze
	//	Returns:
	//		[]*types.Location
	//		[]*types.Motion
	//		[]*time.Time
	//		error
	ParseOutGPSMetadata(pathToVideo string) ([]*types.Location, []*types.Motion, []time.Time, error)
}

// In the form "<deg> deg <deg_mins>' <deg_sconds>\"".
func parseDegrees(degreesStr string) (float64, float64, float64, error) {
	degreesStrSplit := strings.Split(degreesStr, " ")
	if len(degreesStrSplit) != 4 {
		return 0, 0, 0, fmt.Errorf("invalid length for lat/long degrees string %q", degreesStr)
	}

	if degreesStrSplit[1] != "deg" {
		return 0, 0, 0, fmt.Errorf("invalid format for latitude string %q", degreesStr)
	}
	degrees, err := strconv.ParseFloat(degreesStrSplit[0], 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("error parsing float from deg in %q", degreesStr)
	}

	if degreesStrSplit[2][len(degreesStrSplit[2])-1:] != "'" {
		return 0, 0, 0, fmt.Errorf("error parsing lat/long degrees string %q, missing degreeMins symbol", degreesStr)
	}
	degreeMins, err := strconv.ParseFloat(degreesStrSplit[2][:len(degreesStrSplit[2])-1], 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("error parsing float from deg mins in %q", degreesStr)
	}

	if degreesStrSplit[3][len(degreesStrSplit[3])-1:] != "\"" {
		return 0, 0, 0, fmt.Errorf("error parsing lat/long degrees string %q, missing degreeSecs symbol - %q != %q", degreesStr, degreesStrSplit[3][len(degreesStrSplit[3])-1:], "\\\"")
	}
	degreeSecs, err := strconv.ParseFloat(degreesStrSplit[3][:len(degreesStrSplit[3])-1], 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("error parsing float from deg secs in %q", degreesStr)
	}

	return degrees, degreeMins, degreeSecs, nil
}

// StringToLatitude converts a string to a *types.Latitude object.
// Params:
//		latString string: string in the form "<deg> deg <deg_mins>' <deg_sconds>\" <[N | S]>"
// Returns:
//		*types.Latitude
//		error
func StringToLatitude(latString string) (*types.Latitude, error) {
	latStringSplit := strings.Split(latString, " ")
	if len(latStringSplit) != 5 {
		return nil, fmt.Errorf("invalid format for longitude string %q", latString)
	}

	degrees, degreeMins, degreeSecs, err := parseDegrees(strings.Join(latStringSplit[:4], " "))
	if err != nil {
		return nil, fmt.Errorf("error parsing degrees from latString - err: %w", err)
	}

	latitude := &types.Latitude{
		Degrees:       degrees,
		DegreeMinutes: degreeMins,
		DegreeSeconds: degreeSecs,
	}

	switch latStringSplit[4] {
	case "N":
		latitude.LatDirection = types.Latitude_NORTH
	case "S":
		latitude.LatDirection = types.Latitude_SOUTH
	default:
		return nil, fmt.Errorf("error parsing latitude direction from %q", latString)
	}

	return latitude, nil
}

// StringToLongitude converts a string to a *types.Longitude object.
// Params:
//		longString string: string in the form "<deg> deg <deg_mins>' <deg_sconds>\" <[E | W]>"
// Returns:
//		*types.Longitude
//		error
func StringToLongitude(longString string) (*types.Longitude, error) {
	longStringSplit := strings.Split(longString, " ")
	if len(longStringSplit) != 5 {
		return nil, fmt.Errorf("invalid format for longitude string %q", longString)
	}

	degrees, degreeMins, degreeSecs, err := parseDegrees(strings.Join(longStringSplit[:4], " "))
	if err != nil {
		return nil, fmt.Errorf("error parsing degrees from longString - err: %w", err)
	}

	longitude := &types.Longitude{
		Degrees:       degrees,
		DegreeMinutes: degreeMins,
		DegreeSeconds: degreeSecs,
	}

	switch longStringSplit[4] {
	case "E":
		longitude.LongDirection = types.Longitude_EAST
	case "W":
		longitude.LongDirection = types.Longitude_WEST
	default:
		return nil, fmt.Errorf("error parsing longitude direction from %q", longString)
	}

	return longitude, nil
}

package mp4

import (
	"fmt"
	"io/ioutil"
	"os"
	"seneca/api/senecaerror"
	st "seneca/api/type"
	"strconv"
	"strings"
)

// TODO: fix these to be more generic

// CreateTempMP4File creates a temp MP4 file with the given name and returns it.
// Params:
//		name string
// Returns:
//		*os.File
//		error
func CreateTempMP4File(name string) (*os.File, error) {
	nameParts := strings.Split(name, ".")

	if nameParts[len(nameParts)-1] != "mp4" {
		return nil, senecaerror.NewBadStateError(fmt.Errorf("file name %q does not end in 'mp4'", name))
	}

	tempNameNoSuffix := strings.Join(nameParts[:len(nameParts)-1], ".")
	tempName := strings.Join([]string{tempNameNoSuffix, "*", "mp4"}, ".")

	tempFile, err := ioutil.TempFile("", tempName)
	if err != nil {
		return nil, senecaerror.NewBadStateError(fmt.Errorf("error creating temp file %q - err: %v", tempName, err))
	}

	return tempFile, nil
}

// CreateLocalMP4File creates a a file at the given path with the given name.
// Params:
//		name string
//		path string
// Returns:
//		*os.File
//		error
func CreateLocalMP4File(name, path string) (*os.File, error) {
	mp4File, err := os.Create(fmt.Sprintf("%s/%s", path, name))
	if err != nil {
		return nil, fmt.Errorf("error creating local file - err: %v", err)
	}

	return mp4File, nil
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

// StringToLatitude converts a string to a *st.Latitude object.
// Params:
//		latString string: string in the form "<deg> deg <deg_mins>' <deg_sconds>\" <[N | S]>"
// Returns:
//		*st.Latitude
//		error
func StringToLatitude(latString string) (*st.Latitude, error) {
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

// StringToLongitude converts a string to a *st.Longitude object.
// Params:
//		longString string: string in the form "<deg> deg <deg_mins>' <deg_sconds>\" <[E | W]>"
// Returns:
//		*st.Longitude
//		error
func StringToLongitude(longString string) (*st.Longitude, error) {
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

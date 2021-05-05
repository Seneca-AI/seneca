package headerparse

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"seneca/api/constants"
	st "seneca/api/type"
	"seneca/internal/client/logging"
	"seneca/internal/util"
)

const (
	exifToolMetadataMainKey = "Main"
	exifStartTimeKey        = "CreateDate"
	exifDurationKey         = "TrackDuration"
	exifGPSLatKey           = "GPSLatitude"
	exifGPSLongKey          = "GPSLongitude"
	exifGPSDateTimeKey      = "GPSDateTime"
	exifGPSSpeedKey         = "GPSSpeed"
	exifGPSSpeedRefKey      = "GPSSpeedRef"
	exifGPSSampleTimeKey    = "SampleTime"
	// time.Parse requires the first arugment to be a string
	// representing what the datetime 15:04 on 1/2/2006 would be.
	// This is the format that exiftool gives.
	timeParserLayout = "2006:01:02 15:04:05"
)

var (
	exifGPSSpeedRefs        = []string{"mph", "km/h"}
	gpsDateTimeParseLayouts = []string{"2006:01:02 15:04:05.000Z", "2006:01:02 15:04:05.00Z"}
)

type ExifMP4Tool struct {
	logger logging.LoggingInterface
}

func NewExifMP4Tool(logger logging.LoggingInterface) *ExifMP4Tool {
	return &ExifMP4Tool{
		logger: logger,
	}
}

// ParseOutRawVideoMetadata extracts *st.RawVideo metadata from the mp4 at the given path.
func (emt *ExifMP4Tool) ParseOutRawVideoMetadata(pathToVideo string) (*st.RawVideo, error) {
	exifRawData, err := runExifCommand(pathToVideo)
	if err != nil {
		return nil, fmt.Errorf("error running exif command: %w", err)
	}

	unprocessedExifData, err := emt.extractData(exifRawData)
	if err != nil {
		return nil, fmt.Errorf("error extracting unprocessed data: %w", err)
	}

	rawVideo := &st.RawVideo{}

	creationTimeMs, err := getCreationTime(unprocessedExifData.startTime)
	if err != nil {
		return nil, fmt.Errorf("error getting creationTimeMs: %w", err)
	}
	rawVideo.CreateTimeMs = creationTimeMs

	durationMs, err := getDurationMs(unprocessedExifData.duration)
	if err != nil {
		return nil, fmt.Errorf("error getting durationMs: %w", err)
	}
	rawVideo.DurationMs = durationMs

	return rawVideo, nil
}

// 	ParseOutGPSMetadata extracts a list of st.Location, st.Motion and time.Time from the video at the given path.
func (emt *ExifMP4Tool) ParseOutGPSMetadata(pathToVideo string) ([]*st.Location, []*st.Motion, []time.Time, error) {
	exifRawData, err := runExifCommand(pathToVideo)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error running exif command: %w", err)
	}

	unprocessedExifData, err := emt.extractData(exifRawData)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error extracting unprocessed data: %w", err)
	}

	locationsMotionsTimes := []locationMotionTime{}
	for _, gpsData := range unprocessedExifData.gpsData {
		lmt, err := getLocationMotionTime(gpsData)
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

func getCreationTime(timeString string) (int64, error) {
	t, err := time.Parse(timeParserLayout, timeString)
	if err != nil {
		return 0, fmt.Errorf("error parsing CreationTime - err: %v", err)
	}

	if t.Equal(time.Unix(0, 0)) {
		return 0, errors.New("creationTime of 0 is not allowed")
	}

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

func populateAccelerations(motions []*st.Motion) {
	if len(motions) == 0 {
		return
	}

	motions[0].AccelerationMphS = 0

	for i := 1; i < len(motions); i++ {
		motions[i].AccelerationMphS = motions[i].VelocityMph - motions[i-1].VelocityMph
	}
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

type unprocessedExifGPSData struct {
	datetime  string
	latitude  string
	longitude string
	speed     float64
	speedRef  string
}

type unprocessedExifData struct {
	startTime string
	duration  string
	gpsData   []*unprocessedExifGPSData
}

func (emt *ExifMP4Tool) extractData(rawData map[string]interface{}) (*unprocessedExifData, error) {
	unprocessedData := &unprocessedExifData{}

	mainDataObj, ok := rawData[exifToolMetadataMainKey]
	if !ok {
		return nil, fmt.Errorf("no %q data", exifToolMetadataMainKey)
	}

	var err error
	unprocessedData.startTime, unprocessedData.duration, err = extractMainMetadata(mainDataObj)
	if err != nil {
		return nil, fmt.Errorf("error extracting %q data - err: %w", exifToolMetadataMainKey, err)
	}

	for _, subMapObj := range rawData {
		subMap, ok := subMapObj.(map[string]interface{})
		if !ok {
			emt.logger.Warning("expected map[string]interface{} ")
		}

		gpsData, err := extractGPSData(subMap)
		if err != nil {
			return nil, fmt.Errorf("error extracting GPS data: %w", err)
		}
		if gpsData != nil {
			unprocessedData.gpsData = append(unprocessedData.gpsData, gpsData)
		}
	}

	unprocessedData.gpsData = removeDuplicates(unprocessedData.gpsData)
	return unprocessedData, nil
}

func removeDuplicates(dataList []*unprocessedExifGPSData) []*unprocessedExifGPSData {
	gpsDateTimes := map[string]bool{}
	newList := []*unprocessedExifGPSData{}
	for _, gpsData := range dataList {
		if gpsDateTimes[gpsData.datetime] {
			continue
		}
		gpsDateTimes[gpsData.datetime] = true
		newList = append(newList, gpsData)
	}
	return newList
}

func extractMainMetadata(obj interface{}) (string, string, error) {
	startTime := ""
	duration := ""

	mainData, ok := obj.(map[string]interface{})
	if !ok {
		return "", "", fmt.Errorf("expected map[string]interface{} for main metadata %q, got %T", exifToolMetadataMainKey, obj)
	}

	startTimeObj, ok := mainData[exifStartTimeKey]
	if !ok {
		return "", "", fmt.Errorf("no %q data", exifStartTimeKey)
	}
	startTime, ok = startTimeObj.(string)
	if !ok {
		return "", "", fmt.Errorf("expected string for %q, got %T", exifStartTimeKey, startTimeObj)
	}

	durationObj, ok := mainData[exifDurationKey]
	if !ok {
		return "", "", fmt.Errorf("no %q data", exifDurationKey)
	}
	duration, ok = durationObj.(string)
	if !ok {
		return "", "", fmt.Errorf("expected string for %q, got %T", exifDurationKey, durationObj)
	}

	return startTime, duration, nil
}

func extractGPSData(gpsMap map[string]interface{}) (*unprocessedExifGPSData, error) {
	gpsData := &unprocessedExifGPSData{}

	gpsKeys := []string{exifGPSDateTimeKey, exifGPSLatKey, exifGPSLongKey, exifGPSSpeedKey, exifGPSSpeedRefKey, exifGPSSampleTimeKey}
	for _, key := range gpsKeys {
		if _, ok := gpsMap[key]; !ok {
			return nil, nil
		}
	}

	datetime, ok := gpsMap[exifGPSDateTimeKey].(string)
	if !ok {
		return nil, fmt.Errorf("expected string for %q, got %T", exifGPSDateTimeKey, gpsMap[exifGPSDateTimeKey])
	}
	gpsData.datetime = datetime

	lat, ok := gpsMap[exifGPSLatKey].(string)
	if !ok {
		return nil, fmt.Errorf("expected string for %q, got %T", exifGPSLatKey, gpsMap[exifGPSLatKey])
	}
	gpsData.latitude = lat

	long, ok := gpsMap[exifGPSLongKey].(string)
	if !ok {
		return nil, fmt.Errorf("expected string for %q, got %T", exifGPSLongKey, gpsMap[exifGPSLongKey])
	}
	gpsData.longitude = long

	speed, ok := gpsMap[exifGPSSpeedKey].(float64)
	if !ok {
		return nil, fmt.Errorf("expected float64 for %q, got %T", exifGPSSpeedKey, gpsMap[exifGPSSpeedKey])
	}
	gpsData.speed = speed

	speedRef, ok := gpsMap[exifGPSSpeedRefKey].(string)
	if !ok {
		return nil, fmt.Errorf("expected string for %q, got %T", exifGPSSpeedRefKey, gpsMap[exifGPSSpeedRefKey])
	}
	found := false
	for _, validRef := range exifGPSSpeedRefs {
		if speedRef == validRef {
			found = true
		}
	}
	if !found {
		return nil, fmt.Errorf("invalid speedRef %q", speedRef)
	}
	gpsData.speedRef = speedRef

	return gpsData, nil
}

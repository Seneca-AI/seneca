package mp4

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strings"
	"time"

	"seneca/api/constants"
	"seneca/api/senecaerror"
	st "seneca/api/type"
	"seneca/internal/util"

	"github.com/barasher/go-exiftool"
)

const (
	exifToolMetadataMainKey        = "Main"
	exifToolMetadataCreateDateKey  = "CreateDate"
	exifToolMetadataDurationKey    = "TrackDuration"
	exifToolMetadataGPSLatKey      = "GPSLatitude"
	exifToolMetadataGPSLongKey     = "GPSLongitude"
	exifToolMetadataGPSDateTimeKey = "GPSDateTime"
	exifToolMetadataGPSSpeedKey    = "GPSSpeed"
	exifToolMetadataGPSSpeedRefKey = "GPSSpeedRef"
	// time.Parse requires the first arugment to be a string
	// representing what the datetime 15:04 on 1/2/2006 would be.
	// This is the format that exiftool gives.
	timeParserLayout       = "2006:01:02 15:04:05"
	gpsDateTimeParseLayout = "2006:01:02 15:04:05.000Z"
)

var (
	exifToolMainMetdataKeys = []string{exifToolMetadataCreateDateKey, exifToolMetadataDurationKey}
)

type ExifMP4Tool struct {
	exiftool *exiftool.Exiftool
}

func NewExifMP4Tool() (*ExifMP4Tool, error) {
	et, err := exiftool.NewExiftool()
	if err != nil {
		return nil, senecaerror.NewBadStateError(fmt.Errorf("error instantiating exiftool - err: %v", err))
	}
	return &ExifMP4Tool{
		exiftool: et,
	}, nil
}

// ParseOutRawVideoMetadata extracts *st.RawVideo metadata from the mp4 at the given path.
func (emt *ExifMP4Tool) ParseOutRawVideoMetadata(pathToVideo string) (*st.RawVideo, error) {
	var err error
	rawVideo := &st.RawVideo{}

	fileInfoList := emt.exiftool.ExtractMetadata(pathToVideo)
	if len(fileInfoList) < 1 {
		return nil, senecaerror.NewUserError("", fmt.Errorf("fileInfoList for %q is empty", pathToVideo), "MP4 is missing metadata.")
	}

	fileInfo := fileInfoList[0]
	if fileInfo.Err != nil {
		return nil, senecaerror.NewBadStateError(fmt.Errorf("error in fileInfo - err: %v", fileInfo.Err))
	}

	if rawVideo.CreateTimeMs, err = getCreationTimeFromFileMetadata(fileInfo); err != nil {
		return nil, fmt.Errorf("error parsing CreationTime - err: %w", err)
	}

	if rawVideo.DurationMs, err = getDurationFromFileMetadata(fileInfo); err != nil {
		return nil, fmt.Errorf("error parsing Duration - err: %w", err)
	}

	return rawVideo, nil
}

// 	ParseOutGPSMetadata extracts a list of st.Location, st.Motion and time.Time from the video at the given path.
func (emt *ExifMP4Tool) ParseOutGPSMetadata(pathToVideo string) ([]*st.Location, []*st.Motion, []time.Time, error) {
	fileInfoList := emt.exiftool.ExtractMetadata(pathToVideo)
	if len(fileInfoList) < 1 {
		return nil, nil, nil, senecaerror.NewUserError("", fmt.Errorf("fileInfoList for %q is empty", pathToVideo), "MP4 is missing metadata.")
	}

	fileMetadata := fileInfoList[0]
	if fileMetadata.Err != nil {
		return nil, nil, nil, senecaerror.NewBadStateError(fmt.Errorf("error in fileInfo - err: %v", fileMetadata.Err))
	}

	locationsMotionsTimes := []*locationMotionTime{}

	for k, v := range fileMetadata.Fields {
		if strings.Contains(k, "Doc") {
			m, ok := v.(map[string]interface{})
			if !ok {
				// TODO: handle this with the logger
				fmt.Printf("Value is of type: %q\n", reflect.TypeOf(v))
				continue
			}

			locationMotionTime, err := getLocationMotionTimeFromFileMetadataMap(m)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("error parsing GPS metadata map - err: %w", err)
			}

			locationsMotionsTimes = append(locationsMotionsTimes, locationMotionTime)
		}
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

func getCreationTimeFromFileMetadata(fileMetadata exiftool.FileMetadata) (int64, error) {
	mainMap, err := getMainMetadata(fileMetadata)
	if err != nil {
		return 0, fmt.Errorf("error constructing mainMap - err: %w", err)
	}

	timeString, ok := mainMap[exifToolMetadataCreateDateKey]
	if !ok {
		return 0, senecaerror.NewUserError("", fmt.Errorf("could not find value for %q in mainMap", exifToolMetadataCreateDateKey), "MP4 is missing CreationTime metadata.")
	}

	t, err := time.Parse(timeParserLayout, timeString)
	if err != nil {
		return 0, senecaerror.NewUserError("", fmt.Errorf("error parsing CreationTime - err: %v", err), "Malformed MP4 metadata.")
	}

	if t.Equal(time.Unix(0, 0)) {
		return 0, senecaerror.NewUserError("", errors.New("creationTime of 0 is not allowed"), "mp4 has CreationTime of 0, which is not allowed.")
	}

	return util.TimeToMilliseconds(t), nil
}

func getDurationFromFileMetadata(fileMetadata exiftool.FileMetadata) (int64, error) {
	mainMap, err := getMainMetadata(fileMetadata)
	if err != nil {
		return 0, fmt.Errorf("error constructing mainMap - err: %w", err)
	}

	durationString, ok := mainMap[exifToolMetadataDurationKey]
	if !ok {
		return 0, senecaerror.NewUserError("", fmt.Errorf("could not find value for %q in mainMap", exifToolMetadataDurationKey), "MP4 is missing Duration metadata.")
	}

	durationString = strings.Replace(durationString, ":", "h", 1)
	durationString = strings.Replace(durationString, ":", "m", 1)
	durationString = durationString + "s"
	duration, err := time.ParseDuration(durationString)
	if err != nil {
		return 0, senecaerror.NewUserError("", fmt.Errorf("error parsing duration - err: %w", err), "Malformed MP4 metadata.")
	}
	if duration.Milliseconds() == 0 {
		return 0, senecaerror.NewUserError("", fmt.Errorf("duration from durationString %q is zero", durationString), "Malformed MP4 metadata.")
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

func getLocationMotionTimeFromFileMetadataMap(m map[string]interface{}) (*locationMotionTime, error) {
	var err error
	var tempErr error
	var latString, longString, dateTimeString, speedRefString string
	lmt := locationMotionTime{
		location: &st.Location{},
		motion:   &st.Motion{},
	}
	if latString, tempErr = interfaceToString(m[exifToolMetadataGPSLatKey]); tempErr != nil {
		err = tempErr
	}
	if longString, tempErr = interfaceToString(m[exifToolMetadataGPSLongKey]); tempErr != nil {
		err = tempErr
	}
	if dateTimeString, tempErr = interfaceToString(m[exifToolMetadataGPSDateTimeKey]); tempErr != nil {
		err = tempErr
	}

	speed, ok := m[exifToolMetadataGPSSpeedKey].(float64)
	if !ok {
		err = fmt.Errorf("error parsing GPS Speed. Want float64, got %T", m[exifToolMetadataGPSSpeedKey])
	}
	if speedRefString, tempErr = interfaceToString(m[exifToolMetadataGPSSpeedRefKey]); tempErr != nil {
		err = tempErr
	}

	if err != nil {
		if err != nil {
			return &lmt, senecaerror.NewUserError("", fmt.Errorf("error parsing GPS Data %v - err: %v", m, err), "MP4 GPS data malformed.")
		}
	}

	err = nil
	var locLat *st.Latitude
	var locLong *st.Longitude
	if locLat, tempErr = StringToLatitude(latString); tempErr != nil {
		err = tempErr
	}
	if locLong, tempErr = StringToLongitude(longString); tempErr != nil {
		err = tempErr
	}
	lmt.location.Lat = locLat
	lmt.location.Long = locLong

	if lmt.gpsTime, tempErr = time.Parse(gpsDateTimeParseLayout, dateTimeString); tempErr != nil {
		err = tempErr
	}

	switch speedRefString {
	case "mph":
		lmt.motion.VelocityMph = speed
	case "km/h":
		lmt.motion.VelocityMph = math.Round(speed / constants.KilometersToMiles)
	default:
		return &lmt, senecaerror.NewUserError("", fmt.Errorf("error parsing GPS Data %v - err: %v", m, err), "Only mph or km/h are supported speeds.")
	}

	if err != nil {
		return &lmt, senecaerror.NewUserError("", fmt.Errorf("error parsing GPS Data %v - err: %v", m, err), "MP4 GPS data malformed.")
	}
	return &lmt, nil
}

func getMainMetadata(fileMetadata exiftool.FileMetadata) (map[string]string, error) {
	mainMapObj, ok := fileMetadata.Fields[exifToolMetadataMainKey]
	if !ok {
		return nil, senecaerror.NewUserError("", fmt.Errorf("could not find main file metadata"), "MP4 metadata is malformed.")
	}

	mainMap, ok := mainMapObj.(map[string]interface{})
	if !ok {
		return nil, senecaerror.NewBadStateError(fmt.Errorf("want type map[string]interface{} for mainMap in file metadata, got %T", mainMapObj))
	}

	mainMapOut := map[string]string{}
	for _, key := range exifToolMainMetdataKeys {
		valueObj, ok := mainMap[key]
		if !ok {
			return nil, senecaerror.NewUserError("", fmt.Errorf("could not find value for key %q in mainMap", key), fmt.Sprintf("MP4 metadata is malformed and missing %q.", key))
		}
		valueString, err := interfaceToString(valueObj)
		if err != nil {
			return nil, senecaerror.NewUserError("", fmt.Errorf("error converting value for key %q to string - err: %v", key, err), fmt.Sprintf("MP4 metadata is malformed at key %q.", key))
		}
		mainMapOut[key] = valueString
	}
	return mainMapOut, nil
}

func interfaceToString(val interface{}) (string, error) {
	s, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("want string, got %T", val)
	}
	return s, nil
}

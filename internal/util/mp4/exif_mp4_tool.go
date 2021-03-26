package mp4

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/barasher/go-exiftool"
)

// TODO: extract GPS data from videos (it is possible with exiftool)

const (
	exifToolMetadataMainKey        = "Main"
	exifToolMetadataCreateDateKey  = "CreateDate"
	exifToolMetadataDurationKey    = "TrackDuration"
	exifToolMetadataGPSLatKey      = "GPSLatitude"
	exifToolMetadataGPSLongKey     = "GPSLongitude"
	exifToolMetadataGPSDateTimeKey = "GPSDateTime"
	// time.Parse requires the first arugment to be a string
	// representing what the datetime 15:04 on 1/2/2006 would be.
	// This is the format that exiftool gives.
	timeParserLayout       = "2006:01:02 15:04:05"
	gpsDateTimeParseLayout = "2006:01:02 15:04:05.000Z"
)

var (
	exifToolMainMetdataKeys = []string{exifToolMetadataCreateDateKey, exifToolMetadataDurationKey}
)

type ExitMP4Tool struct {
	exiftool *exiftool.Exiftool
}

func NewExitMP4Tool() (*ExitMP4Tool, error) {
	et, err := exiftool.NewExiftool()
	if err != nil {
		return nil, fmt.Errorf("error instantiating exiftool - err: %v", err)
	}
	return &ExitMP4Tool{
		exiftool: et,
	}, nil
}

// GetMetadata extracts VideoMetadata from the video at the given path.
func (emt *ExitMP4Tool) GetMetadata(pathToVideo string) (*VideoMetadata, error) {
	var err error
	videoMetadata := &VideoMetadata{}

	fileInfoList := emt.exiftool.ExtractMetadata(pathToVideo)
	if len(fileInfoList) < 1 {
		return nil, fmt.Errorf("fileInfoList for %q is empty", pathToVideo)
	}

	fileInfo := fileInfoList[0]
	if fileInfo.Err != nil {
		return nil, fmt.Errorf("error in fileInfo - err: %v", fileInfo.Err)
	}

	if videoMetadata.CreationTime, err = getCreationTimeFromFileMetadata(fileInfo); err != nil {
		return nil, fmt.Errorf("error parsing CreationTime - err: %v", err)
	}

	if videoMetadata.Duration, err = getDurationFromFileMetadata(fileInfo); err != nil {
		return nil, fmt.Errorf("error parsing Duration - err: %v", err)
	}

	return videoMetadata, nil
}

func getCreationTimeFromFileMetadata(fileMetadata exiftool.FileMetadata) (*time.Time, error) {
	mainMap, err := getMainMetadata(fileMetadata)
	if err != nil {
		return nil, fmt.Errorf("error constructing mainMap - err: %v", err)
	}

	timeString, ok := mainMap[exifToolMetadataCreateDateKey]
	if !ok {
		return nil, fmt.Errorf("could not find value for %q in mainMap", exifToolMetadataCreateDateKey)
	}

	t, err := time.Parse(timeParserLayout, timeString)
	if err != nil {
		return nil, fmt.Errorf("error parsing CreationTime - err: %v", err)
	}
	return &t, nil
}

func getDurationFromFileMetadata(fileMetadata exiftool.FileMetadata) (*time.Duration, error) {
	mainMap, err := getMainMetadata(fileMetadata)
	if err != nil {
		return nil, fmt.Errorf("error constructing mainMap - err: %v", err)
	}

	durationString, ok := mainMap[exifToolMetadataDurationKey]
	if !ok {
		return nil, fmt.Errorf("could not find value for %q in mainMap", exifToolMetadataDurationKey)
	}

	durationString = strings.Replace(durationString, ":", "h", 1)
	durationString = strings.Replace(durationString, ":", "m", 1)
	durationString = durationString + "s"
	duration, err := time.ParseDuration(durationString)
	if err != nil {
		return nil, fmt.Errorf("error parsing Duration - err: %v", err)
	}
	return &duration, nil
}

func getLocationsFromFileMetadata(fileMetadata exiftool.FileMetadata) ([]Location, error) {
	locations := []Location{}
	var err error
	for k, v := range fileMetadata.Fields {
		if strings.Contains(k, "Doc") {
			m, ok := v.(map[string]interface{})
			if !ok {
				fmt.Printf("value is of type: %q\n", reflect.TypeOf(v))
				continue
			}
			location := Location{}

			var tempErr error
			if location.lat, tempErr = interfaceToString(m[exifToolMetadataGPSLatKey]); tempErr != nil {
				err = tempErr
			}
			if location.long, tempErr = interfaceToString(m[exifToolMetadataGPSLongKey]); tempErr != nil {
				err = tempErr
			}
			var dateTimeString string
			if dateTimeString, tempErr = interfaceToString(m[exifToolMetadataGPSDateTimeKey]); tempErr != nil {
				err = tempErr
			}
			if location.timestamp, tempErr = time.Parse(gpsDateTimeParseLayout, dateTimeString); tempErr != nil {
				err = tempErr
			}

			if err != nil {
				return nil, fmt.Errorf("error parsing GPS Data - err: %v", err)
			}

			locations = append(locations, location)
		}
	}
	sort.Slice(locations, func(i, j int) bool { return locations[i].timestamp.Before(locations[j].timestamp) })

	return locations, nil
}

func getMainMetadata(fileMetadata exiftool.FileMetadata) (map[string]string, error) {
	mainMapObj, ok := fileMetadata.Fields[exifToolMetadataMainKey]
	if !ok {
		return nil, fmt.Errorf("could not find main file metadata")
	}

	mainMap, ok := mainMapObj.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("want type map[string]interface{} for mainMap in file metadata, got %T", mainMapObj)
	}

	mainMapOut := map[string]string{}
	for _, key := range exifToolMainMetdataKeys {
		valueObj, ok := mainMap[key]
		if !ok {
			return nil, fmt.Errorf("could not find value for key %q in mainMap", key)
		}
		valueString, err := interfaceToString(valueObj)
		if err != nil {
			return nil, fmt.Errorf("error converting value for key %q to string - err: %v", key, err)
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

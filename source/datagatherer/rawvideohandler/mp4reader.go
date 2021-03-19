package rawvideohandler

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/barasher/go-exiftool"
)

// TODO: extract GPS data from videos (it is possible with exiftool)

const (
	createDateKey = "CreateDate"
	durationKey   = "TrackDuration"
	// time.Parse requires the first arugment to be a string
	// representing what the datetime 15:04 on 1/2/2006 would be.
	// This is the format that exiftool gives.
	timeParserLayout = "2006:01:02 15:04:05"
)

type VideoMetadata struct {
	CreationTime time.Time
	Duration     time.Duration
}

func (vmd *VideoMetadata) String() string {
	// TODO: handle this error with a logger
	b, _ := json.MarshalIndent(vmd, "", "\t")
	return fmt.Sprint(string(b))
}

// GetMetadata extracts VideoMetadata from the video at the given path.
// Params:
// 		string pathToVideo: path to video to get metadata from
// Returns:
// 		*VideoMetadata: the VideoMetadata object
//		error
func GetMetadata(pathToVideo string) (*VideoMetadata, error) {
	fmt.Printf("Getting metadata for video %q\n", pathToVideo)

	videoMetadata := &VideoMetadata{}

	et, err := exiftool.NewExiftool()
	if err != nil {
		return nil, fmt.Errorf("error instantiating exiftool - err: %v", err)
	}

	fileInfoList := et.ExtractMetadata(pathToVideo)
	if len(fileInfoList) < 1 {
		return nil, fmt.Errorf("fileInfoList for %q is empty", pathToVideo)
	}

	fileInfo := fileInfoList[0]
	if fileInfo.Err != nil {
		return nil, fmt.Errorf("error in fileInfo - err: %v", fileInfo.Err)
	}

	for k, _ := range fileInfo.Fields {
		value, err := fileInfo.GetString(k)
		if err != nil {
			return nil, fmt.Errorf("unknown key in fileInfo %s", k)
		}
		if k == createDateKey {
			t, err := time.Parse(timeParserLayout, value)
			if err != nil {
				return nil, fmt.Errorf("error parsing CreationTime - err: %v", err)
			}
			videoMetadata.CreationTime = t
		}
		if k == durationKey {
			durationString := strings.Replace(value, ":", "h", 1)
			durationString = strings.Replace(durationString, ":", "m", 1)
			durationString = durationString + "s"
			duration, err := time.ParseDuration(durationString)
			if err != nil {
				return nil, fmt.Errorf("error parsing Duration - err: %v", err)
			}
			videoMetadata.Duration = duration
		}
	}
	return videoMetadata, nil
}

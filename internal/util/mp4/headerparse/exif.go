// Package headerparse extracts MP4 metadata by:
//		1. 	Running exiftool (exif_wrapper.go)
//		2. 	Extracting the data into unstructured string values (exif_extract.go)
//		3.	Parsing the data into real values (exif_parse.go)
//		4.	Putting it all together into Seneca types (exif.go)
//
//	Some notes:
//		1. All times should be in UTC rooted at the video's 'CreateTime' (or similar field).
package headerparse

// TODO(lucaloncar): run a bunch of other videos through to make sure timestamps aren't messed up

import (
	"fmt"
	st "seneca/api/type"
	"seneca/internal/client/logging"
	"sort"
	"time"
)

const (
	exifToolMetadataMainKey = "Main"
	exifDurationKey         = "TrackDuration"
	exifGPSLatKey           = "GPSLatitude"
	exifGPSLongKey          = "GPSLongitude"
	exifGPSSpeedKey         = "GPSSpeed"
	exifGPSSpeedRefKey      = "GPSSpeedRef"
	exifGPSSampleTimeKey    = "SampleTime"
)

type parserValues struct {
	videoStartTimeKey string
	// time.Parse requires the first arugment to be a string
	// representing what the datetime 15:04 on 1/2/2006 would be.
	videoStartTimeLayout string
	gpsTimeKey           string
	gpsTimeLayout        string
	gpsTimeAdjustment    time.Duration
}

type DashCamName string

var (
	Garmin55          DashCamName = "Garmin55"
	BlackVueDR750X1CH DashCamName = "BlackVue_DR750X-1CH"
)

func (dcn DashCamName) String() string {
	return string(dcn)
}

var (
	exifGPSSpeedRefs        = []string{"mph", "km/h"}
	gpsDateTimeParseLayouts = []string{"2006:01:02 15:04:05.000Z", "2006:01:02 15:04:05.00Z"}

	cameraDataLayouts = map[DashCamName]*parserValues{
		Garmin55: {
			videoStartTimeKey:    "CreateDate",
			videoStartTimeLayout: "2006:01:02 15:04:05",
			gpsTimeKey:           "GPSDateTime",
			gpsTimeLayout:        "2006:01:02 15:04:05.000Z",
		},
		BlackVueDR750X1CH: {
			videoStartTimeKey:    "StartTime",
			videoStartTimeLayout: "2006:01:02 15:04:05.000",
			gpsTimeKey:           "GPSDateTime",
			gpsTimeLayout:        "2006:01:02 15:04:05.00Z",
			gpsTimeAdjustment:    (time.Second * 2),
		},
	}
)

type ExifMP4Tool struct {
	logger logging.LoggingInterface
}

func NewExifMP4Tool(logger logging.LoggingInterface) *ExifMP4Tool {
	return &ExifMP4Tool{
		logger: logger,
	}
}

func (emt *ExifMP4Tool) ParseVideoMetadata(pathToVideo string) (*st.RawVideo, []*st.Location, []*st.Motion, []time.Time, error) {
	exifRawData, err := runExifCommand(pathToVideo)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("error running exif command: %w", err)
	}

	dashCamName, unprocessedExifData, err := emt.extractData(exifRawData)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("error extracting unprocessed data: %w", err)
	}

	rawVideo, err := parseOutRawVideoMetadata(dashCamName, unprocessedExifData)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("error parsing rawVideo metadata: %w", err)
	}

	locations, motions, times, err := parseOutGPSMetadata(unprocessedExifData)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("error parsing location/motion metadata: %w", err)
	}

	times, err = adjustTimestamps(dashCamName, rawVideo, times)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	if err := validateData(rawVideo, locations, motions, times); err != nil {
		return nil, nil, nil, nil, err
	}

	return rawVideo, locations, motions, times, nil
}

// parseOutRawVideoMetadata extracts *st.RawVideo metadata from the mp4 at the given path.
func parseOutRawVideoMetadata(dashCamName DashCamName, unprocessedExifData *unprocessedExifData) (*st.RawVideo, error) {
	rawVideo := &st.RawVideo{}

	creationTimeMs, err := getCreationTime(dashCamName, unprocessedExifData.startTime)
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

// 	parseOutGPSMetadata extracts a list of st.Location, st.Motion and time.Time from the video at the given path.
func parseOutGPSMetadata(unprocessedExifData *unprocessedExifData) ([]*st.Location, []*st.Motion, []time.Time, error) {
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

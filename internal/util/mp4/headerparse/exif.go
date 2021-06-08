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
	"time"
)

type DashCamName string

var (
	Garmin55          DashCamName = "Garmin55"
	BlackVueDR750X1CH DashCamName = "BlackVue_DR750X-1CH"
)

func (dcn DashCamName) String() string {
	return string(dcn)
}

var (
	exifGPSSpeedRefs = []string{"mph", "km/h"}
)

type exifParserInterface interface {
	init(unprocessedExifData *unprocessedExifData)
	parseOutRawVideoMetadata() (*st.RawVideo, error)
	parseOutGPSMetadata(rawVideo *st.RawVideo) ([]*st.Location, []*st.Motion, []time.Time, error)
}

type ExifMP4Tool struct {
	logger  logging.LoggingInterface
	parsers map[DashCamName]exifParserInterface
}

func NewExifMP4Tool(logger logging.LoggingInterface) *ExifMP4Tool {
	return &ExifMP4Tool{
		logger: logger,
		parsers: map[DashCamName]exifParserInterface{
			BlackVueDR750X1CH: &blackVueDR750X1CHExifParser{},
			Garmin55:          &Garmin55ExifParser{},
		},
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

	parser := emt.parsers[dashCamName]
	parser.init(unprocessedExifData)

	rawVideo, err := parser.parseOutRawVideoMetadata()
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("error parsing rawVideo metadata: %w", err)
	}

	locations, motions, times, err := parser.parseOutGPSMetadata(rawVideo)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("error parsing location/motion metadata: %w", err)
	}

	if err := validateData(rawVideo, locations, motions, times); err != nil {
		return nil, nil, nil, nil, err
	}

	return rawVideo, locations, motions, times, nil
}

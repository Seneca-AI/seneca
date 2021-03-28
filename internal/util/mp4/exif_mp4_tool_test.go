package mp4

import (
	"errors"
	"seneca/api/senecaerror"
	"seneca/api/types"
	"seneca/internal/util"
	"testing"
	"time"
)

func TestGetMetadataHasExpectedData(t *testing.T) {
	if util.IsCIEnv() {
		t.Skip("Skipping exiftool test in GitHub env.")
	}

	exifMP4Tool, err := NewExitMP4Tool()
	if err != nil {
		t.Errorf("NewExitMP4Tool() returns err: %v", err)
	}

	pathToTestMp4 := "../../../test/testdata/dad_example.MP4"
	expectedCreationTimeMs := util.TimeToMilliseconds(time.Date(2021, time.February, 13, 17, 47, 49, 0, time.UTC))
	expectedDurationMs := util.DurationToMilliseconds(time.Minute)

	rawVideo, err := exifMP4Tool.ParseOutRawVideoMetadata(pathToTestMp4)
	if err != nil {
		t.Errorf("GetMetadata(%s) returns err: %v", pathToTestMp4, err)
		return
	}

	if rawVideo.GetCreateTimeMs() != expectedCreationTimeMs {
		t.Errorf("rawVideo.CreateTimeMs incorrect. got %v, want %v", rawVideo.CreateTimeMs, expectedCreationTimeMs)
	}
	if rawVideo.GetDurationMs() != expectedDurationMs {
		t.Errorf("rawVideo.GetDurationMs incorrect. got %v, want %v", rawVideo.GetDurationMs(), expectedDurationMs)
	}
}

func TestGetMetadataDoesntCrashWitoutVideoFile(t *testing.T) {
	if util.IsCIEnv() {
		t.Skip("Skipping exiftool test in GitHub env.")
	}

	exifMP4Tool, err := NewExitMP4Tool()
	if err != nil {
		t.Errorf("NewExitMP4Tool() returns err: %v", err)
	}

	_, err = exifMP4Tool.ParseOutRawVideoMetadata("../idontexist")
	if err == nil {
		t.Errorf("Want non-nil error from bogus input file, got nil")
	}
	var bse *senecaerror.BadStateError
	if !errors.As(err, &bse) {
		t.Errorf("Want BadStateError, got %v", err)
	}
}

func TestGetLocationMotionTimeFromFileMetadataMap(t *testing.T) {
	goodInputMap := map[string]interface{}{
		"GPSLatitude":  "40 deg 24' 55.86\" N",
		"GPSLongitude": "74 deg 25' 50.17\" W",
		"GPSSpeed":     float64(41),
		"GPSSpeedRef":  "mph",
		"GPSDateTime":  "2021:02:13 22:48:47.000Z",
	}

	goodOutputStruct := &locationMotionTime{
		location: &types.Location{
			Lat: &types.Latitude{
				Degrees:       40,
				DegreeMinutes: 24,
				DegreeSeconds: 55.86,
				LatDirection:  types.Latitude_NORTH,
			},
			Long: &types.Longitude{
				Degrees:       74,
				DegreeMinutes: 25,
				DegreeSeconds: 50.17,
				LongDirection: types.Longitude_WEST,
			},
		},
		motion: &types.Motion{
			VelocityMph: 41,
		},
		gpsTime: time.Date(2021, 02, 13, 22, 48, 47, 0, time.UTC),
	}

	goodInputMapKmh := map[string]interface{}{
		"GPSLatitude":  "40 deg 24' 55.86\" N",
		"GPSLongitude": "74 deg 25' 50.17\" W",
		"GPSSpeed":     float64(41),
		"GPSSpeedRef":  "km/h",
		"GPSDateTime":  "2021:02:13 22:48:47.000Z",
	}
	goodOutputStructKmh := &locationMotionTime{
		location: &types.Location{
			Lat: &types.Latitude{
				Degrees:       40,
				DegreeMinutes: 24,
				DegreeSeconds: 55.86,
				LatDirection:  types.Latitude_NORTH,
			},
			Long: &types.Longitude{
				Degrees:       74,
				DegreeMinutes: 25,
				DegreeSeconds: 50.17,
				LongDirection: types.Longitude_WEST,
			},
		},
		motion: &types.Motion{
			VelocityMph: 65.98312441359509,
		},
		gpsTime: time.Date(2021, 02, 13, 22, 48, 47, 0, time.UTC),
	}

	inputMapUnsupportedSpeedRef := map[string]interface{}{
		"GPSLatitude":  "40 deg 24' 55.86\" N",
		"GPSLongitude": "74 deg 25' 50.17\" W",
		"GPSSpeed":     41,
		"GPSSpeedRef":  "q",
		"GPSDateTime":  "2021:02:13 22:48:47.000Z",
	}

	inputMapBadTimeFormat := map[string]interface{}{
		"GPSLatitude":  "40 deg 24' 55.86\" N",
		"GPSLongitude": "74 deg 25' 50.17\" W",
		"GPSSpeed":     41,
		"GPSSpeedRef":  "mph",
		"GPSDateTime":  "02/13/2021 22:48:47.000Z",
	}

	inputMapBadSpeedType := map[string]interface{}{
		"GPSLatitude":  "40 deg 24' 55.86\" N",
		"GPSLongitude": "74 deg 25' 50.17\" W",
		"GPSSpeed":     "41",
		"GPSSpeedRef":  "mph",
		"GPSDateTime":  "02/13/2021 22:48:47.000Z",
	}

	testCases := []struct {
		desc     string
		inputMap map[string]interface{}
		want     *locationMotionTime
		wantErr  bool
	}{
		{
			desc:     "test empty map throws error",
			inputMap: make(map[string]interface{}),
			want:     nil,
			wantErr:  true,
		},
		{
			desc:     "test expected output",
			inputMap: goodInputMap,
			want:     goodOutputStruct,
			wantErr:  false,
		},
		{
			desc:     "test expected output kmh",
			inputMap: goodInputMapKmh,
			want:     goodOutputStructKmh,
			wantErr:  false,
		},
		{
			desc:     "test bad speed ref returns err",
			inputMap: inputMapUnsupportedSpeedRef,
			want:     nil,
			wantErr:  true,
		},
		{
			desc:     "test bad time format returns err",
			inputMap: inputMapBadTimeFormat,
			want:     nil,
			wantErr:  true,
		},
		{
			desc:     "test bad speed type returns err",
			inputMap: inputMapBadSpeedType,
			want:     nil,
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := getLocationMotionTimeFromFileMetadataMap(tc.inputMap)
			if tc.wantErr {
				if err == nil {
					t.Errorf("Want err from getLocationMotionTimeFromFileMetadataMap(%v), got nil", tc.inputMap)
				}
				return
			}
			if err != nil {
				t.Errorf("getLocationMotionTimeFromFileMetadataMap(%v) returns unexpected err %v", tc.inputMap, err)
				return
			}
			if !util.LocationsEquals(tc.want.location, got.location) {
				t.Errorf("Locations not equal. Got %v, want %v", got.location, tc.want.location)
			}
			if got.motion.VelocityMph != tc.want.motion.VelocityMph || got.motion.AccelerationMphS != tc.want.motion.AccelerationMphS {
				t.Errorf("Motions not equal.  Got %v, want %v", got.motion, tc.want.motion)
			}
			if got.gpsTime != tc.want.gpsTime {
				t.Errorf("Times not equal. Got %v, want %v.", got.gpsTime, tc.want.gpsTime)
			}
		})
	}
}

func TestGetLocationDataFileMetadata(t *testing.T) {
	if util.IsCIEnv() {
		t.Skip("Skipping exiftool test in GitHub env.")
	}

	exifTool, err := NewExitMP4Tool()
	if err != nil {
		t.Error(err)
	}

	pathToTestMp4 := "../../../test/testdata/dad_example.MP4"
	fileInfoList := exifTool.exiftool.ExtractMetadata(pathToTestMp4)
	if len(fileInfoList) < 1 {
		t.Errorf("FileInfoList for %q is empty", pathToTestMp4)
	}

	fileInfo := fileInfoList[0]
	if fileInfo.Err != nil {
		t.Errorf("Error in fileInfo - err: %v", fileInfo.Err)
	}

	expectedAccelerations := []float64{0, -2, -2, -2, -2, -3, -1, -2, 0, -5, -2, -2, 0, 3, 4, 2, 3, 2, 1, 0, 0, -1, -1, -3, -2, -3, -2, 0, 0, -1, 1, 2, 1, 2, 1, 1, 1, 9, 2, 1, 0, 0, 1, 1, 1, 0, 1, 2, 1, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0}

	_, motions, _, err := getLocationDataFileMetadata(fileInfoList[0])
	if err != nil {
		t.Errorf("getLocationDataFileMetadata() returns err - %v", err)
	}

	if len(motions) != len(expectedAccelerations) {
		t.Errorf("Want len %d for motions, got %d", len(expectedAccelerations), len(motions))
	}

	for i, m := range motions {
		if expectedAccelerations[i] != m.AccelerationMphS {
			t.Errorf("Want accelerations %f at index %d, got %f", expectedAccelerations[i], i, m.AccelerationMphS)
		}
	}
}

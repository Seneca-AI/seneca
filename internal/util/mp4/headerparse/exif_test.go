package headerparse

import (
	"errors"
	"seneca/api/senecaerror"
	st "seneca/api/type"
	"seneca/internal/client/logging"
	"seneca/internal/util"
	"testing"
	"time"
)

func TestGetMetadataHasExpectedData(t *testing.T) {
	if util.IsCIEnv() {
		t.Skip("Skipping exiftool test in GitHub env.")
	}

	testCases := []struct {
		desc             string
		pathToVideo      string
		wantCreateTimeMs int64
		wantDurationMs   int64
	}{
		{
			desc:             "garmin",
			pathToVideo:      "../../../../test/testdata/garmin_example.mp4",
			wantCreateTimeMs: util.TimeToMilliseconds(time.Date(2021, time.February, 13, 17, 47, 49, 0, time.UTC)),
			wantDurationMs:   time.Minute.Milliseconds(),
		},
		{
			desc:             "blackvue",
			pathToVideo:      "../../../../test/testdata/blackvue_example.mp4",
			wantCreateTimeMs: util.TimeToMilliseconds(time.Date(2021, time.April, 4, 21, 50, 29, 0, time.UTC)),
			wantDurationMs:   time.Minute.Milliseconds(),
		},
	}

	exifMP4Tool := NewExifMP4Tool(logging.NewLocalLogger(false))

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			rawVideo, err := exifMP4Tool.ParseOutRawVideoMetadata(tc.pathToVideo)
			if err != nil {
				t.Errorf("ParseOutRawVideoMetadata(%s) returns err: %v", tc.pathToVideo, err)
				return
			}

			if rawVideo.GetCreateTimeMs() != tc.wantCreateTimeMs {
				t.Errorf("rawVideo.CreateTimeMs incorrect. got %v, want %v", util.MillisecondsToTime(rawVideo.CreateTimeMs), util.MillisecondsToTime(tc.wantCreateTimeMs))
			}
			if rawVideo.GetDurationMs() != tc.wantDurationMs {
				t.Errorf("rawVideo.GetDurationMs incorrect. got %v, want %v", rawVideo.GetDurationMs(), tc.wantDurationMs)
			}
		})
	}
}

func TestGetMetadataHasRejectsFileWithZeroCreationTime(t *testing.T) {
	if util.IsCIEnv() {
		t.Skip("Skipping exiftool test in GitHub env.")
	}

	exifMP4Tool := NewExifMP4Tool(logging.NewLocalLogger(false))

	pathToNoMetadataMP4 := "../../../../test/testdata/no_metadata.mp4"
	_, err := exifMP4Tool.ParseOutRawVideoMetadata(pathToNoMetadataMP4)
	if err == nil {
		t.Errorf("Expected err from exifMP4Tool.ParseOutRawVideoMetadata(%q), got nil", pathToNoMetadataMP4)
	}
}

func TestGetMetadataDoesntCrashWitoutVideoFile(t *testing.T) {
	if util.IsCIEnv() {
		t.Skip("Skipping exiftool test in GitHub env.")
	}

	exifMP4Tool := NewExifMP4Tool(logging.NewLocalLogger(false))

	_, err := exifMP4Tool.ParseOutRawVideoMetadata("../idontexist")
	if err == nil {
		t.Errorf("Want non-nil error from bogus input file, got nil")
	}
	var bse *senecaerror.BadStateError
	if !errors.As(err, &bse) {
		t.Errorf("Want BadStateError, got %v", err)
	}
}

func TestGetLocationMotionTime(t *testing.T) {
	goodUnprocessedGPSDataMPH := &unprocessedExifGPSData{
		datetime:  "2021:02:13 22:48:47.000Z",
		latitude:  "40 deg 24' 55.86\" N",
		longitude: "74 deg 25' 50.17\" W",
		speed:     float64(41),
		speedRef:  "mph",
	}
	goodOutputStruct := &locationMotionTime{
		location: &st.Location{
			Lat: &st.Latitude{
				Degrees:       40,
				DegreeMinutes: 24,
				DegreeSeconds: 55.86,
				LatDirection:  st.Latitude_NORTH,
			},
			Long: &st.Longitude{
				Degrees:       74,
				DegreeMinutes: 25,
				DegreeSeconds: 50.17,
				LongDirection: st.Longitude_WEST,
			},
		},
		motion: &st.Motion{
			VelocityMph: 41,
		},
		gpsTime: time.Date(2021, 02, 13, 22, 48, 47, 0, time.UTC),
	}

	goodUnprocessedGPSDataKMH := &unprocessedExifGPSData{
		datetime:  "2021:02:13 22:48:47.000Z",
		latitude:  "40 deg 24' 55.86\" N",
		longitude: "74 deg 25' 50.17\" W",
		speed:     float64(41),
		speedRef:  "km/h",
	}
	goodOutputStructKMH := &locationMotionTime{
		location: &st.Location{
			Lat: &st.Latitude{
				Degrees:       40,
				DegreeMinutes: 24,
				DegreeSeconds: 55.86,
				LatDirection:  st.Latitude_NORTH,
			},
			Long: &st.Longitude{
				Degrees:       74,
				DegreeMinutes: 25,
				DegreeSeconds: 50.17,
				LongDirection: st.Longitude_WEST,
			},
		},
		motion: &st.Motion{
			VelocityMph: 25,
		},
		gpsTime: time.Date(2021, 02, 13, 22, 48, 47, 0, time.UTC),
	}

	unsupportedSpeedRefData := &unprocessedExifGPSData{
		datetime:  "2021:02:13 22:48:47.000Z",
		latitude:  "40 deg 24' 55.86\" N",
		longitude: "74 deg 25' 50.17\" W",
		speed:     float64(41),
		speedRef:  "q",
	}

	badTimeFormatData := &unprocessedExifGPSData{
		datetime:  "02/13/2021 22:48:47.000Z",
		latitude:  "40 deg 24' 55.86\" N",
		longitude: "74 deg 25' 50.17\" W",
		speed:     float64(41),
		speedRef:  "mph",
	}

	testCases := []struct {
		desc    string
		input   *unprocessedExifGPSData
		want    *locationMotionTime
		wantErr bool
	}{
		{
			desc:    "test expected output",
			input:   goodUnprocessedGPSDataMPH,
			want:    goodOutputStruct,
			wantErr: false,
		},
		{
			desc:    "test expected output kmh",
			input:   goodUnprocessedGPSDataKMH,
			want:    goodOutputStructKMH,
			wantErr: false,
		},
		{
			desc:    "test bad speed ref returns err",
			input:   unsupportedSpeedRefData,
			want:    nil,
			wantErr: true,
		},
		{
			desc:    "test bad time format returns err",
			input:   badTimeFormatData,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := getLocationMotionTime(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Errorf("Want err from getLocationMotionTime(%v), got nil", tc.input)
				}
				return
			}
			if err != nil {
				t.Errorf("getLocationMotionTime(%v) returns unexpected err %v", tc.input, err)
				return
			}
			if !util.LocationsEqual(tc.want.location, got.location) {
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

func TestParseOutGPSMetadata(t *testing.T) {
	if util.IsCIEnv() {
		t.Skip("Skipping exiftool test in GitHub env.")
	}

	testCases := []struct {
		desc                  string
		pathToVideo           string
		expectedAccelerations []float64
	}{
		{
			desc:                  "garmin",
			pathToVideo:           "../../../../test/testdata/garmin_example.mp4",
			expectedAccelerations: []float64{0, -2, -2, -2, -2, -3, -1, -2, 0, -5, -2, -2, 0, 3, 4, 2, 3, 2, 1, 0, 0, -1, -1, -3, -2, -3, -2, 0, 0, -1, 1, 2, 1, 2, 1, 1, 1, 9, 2, 1, 0, 0, 1, 1, 1, 0, 1, 2, 1, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0},
		},
		// TODO(lucaloncar): double check this
		{
			desc:                  "blackvue",
			pathToVideo:           "../../../../test/testdata/blackvue_example.mp4",
			expectedAccelerations: []float64{0, 0, 0, 0, 0, -1, 0, 0, 0, 0, 0, 0, -1, -1, -2, -1, 0, 0, -1, 0, 0, 0, 1, 0, 0, 0, -1, -1, 3, -2, 0, 0, 0, 0, 0, 0, 0, 1, 0, 1, -1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, -1, 0, 1, 0, 0},
		},
	}

	exifTool := NewExifMP4Tool(logging.NewLocalLogger(false))

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			_, motions, _, err := exifTool.ParseOutGPSMetadata(tc.pathToVideo)
			if err != nil {
				t.Errorf("ParseOutGPSMetadata() returns err - %v", err)
			}

			if len(motions) != len(tc.expectedAccelerations) {
				t.Errorf("Want len %d for motions, got %d", len(tc.expectedAccelerations), len(motions))
				return
			}

			for i, m := range motions {
				if tc.expectedAccelerations[i] != m.AccelerationMphS {
					t.Errorf("Want accelerations %f at index %d, got %f", tc.expectedAccelerations[i], i, m.AccelerationMphS)
					t.Errorf("%v", motions)
				}
			}
		})
	}

}

func TestStringToLatitude(t *testing.T) {
	testCases := []struct {
		desc     string
		inputStr string
		want     *st.Latitude
		wantErr  bool
	}{
		{
			desc:     "test empty string throws err",
			inputStr: "",
			want:     nil,
			wantErr:  true,
		},
		{
			desc:     "test malformed string without direction throws err",
			inputStr: "40 deg 24' 57.66\"",
			want:     nil,
			wantErr:  true,
		},
		{
			desc:     "test malformed string with extra info throws err",
			inputStr: "40 deg 24' 57.66\" N P",
			want:     nil,
			wantErr:  true,
		},
		{
			desc:     "test malformed string with bad degrees throws err",
			inputStr: "40 degrees 24' 57.66\" N",
			want:     nil,
			wantErr:  true,
		},
		{
			desc:     "test malformed string without degrees symbol throws err",
			inputStr: "40 deg 24 57.66 N",
			want:     nil,
			wantErr:  true,
		},
		{
			desc:     "test string without decimals succeeds",
			inputStr: "40 deg 24' 57\" N",
			want: &st.Latitude{
				Degrees:       40,
				DegreeMinutes: 24,
				DegreeSeconds: 57,
				LatDirection:  st.Latitude_NORTH,
			},
			wantErr: false,
		},
		{
			desc:     "test string with all decimals and south succeeds",
			inputStr: "40.35 deg 24.56' 57.78\" S",
			want: &st.Latitude{
				Degrees:       40.35,
				DegreeMinutes: 24.56,
				DegreeSeconds: 57.78,
				LatDirection:  st.Latitude_SOUTH,
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := stringToLatitude(tc.inputStr)
			if tc.wantErr {
				if err == nil {
					t.Errorf("Want err from stringToLatitude(%s), got nil", tc.inputStr)
				}
				return
			}
			if err != nil {
				t.Errorf("stringToLatitude(%s) returns unexpected err: %v", tc.inputStr, err)
				return
			}
			if got.Degrees != tc.want.Degrees || got.DegreeMinutes != tc.want.DegreeMinutes || got.DegreeSeconds != tc.want.DegreeSeconds || got.LatDirection != tc.want.LatDirection {
				t.Errorf("Want %v from stringToLatitude(%s), got %v", tc.want, tc.inputStr, got)
			}
		})
	}
}

func TestStringToLongitude(t *testing.T) {
	testCases := []struct {
		desc     string
		inputStr string
		want     *st.Longitude
		wantErr  bool
	}{
		{
			desc:     "test empty string throws err",
			inputStr: "",
			want:     nil,
			wantErr:  true,
		},
		{
			desc:     "test malformed string without direction throws err",
			inputStr: "40 deg 24' 57.66\"",
			want:     nil,
			wantErr:  true,
		},
		{
			desc:     "test malformed string with extra info throws err",
			inputStr: "40 deg 24' 57.66\" E P",
			want:     nil,
			wantErr:  true,
		},
		{
			desc:     "test malformed string with bad degrees throws err",
			inputStr: "40 degrees 24' 57.66\" E",
			want:     nil,
			wantErr:  true,
		},
		{
			desc:     "test malformed string without degrees symbol throws err",
			inputStr: "40 deg 24 57.66 E",
			want:     nil,
			wantErr:  true,
		},
		{
			desc:     "test string without decimals succeeds",
			inputStr: "40 deg 24' 57\" E",
			want: &st.Longitude{
				Degrees:       40,
				DegreeMinutes: 24,
				DegreeSeconds: 57,
				LongDirection: st.Longitude_EAST,
			},
			wantErr: false,
		},
		{
			desc:     "test string with all decimals and west succeeds",
			inputStr: "40.35 deg 24.56' 57.78\" W",
			want: &st.Longitude{
				Degrees:       40.35,
				DegreeMinutes: 24.56,
				DegreeSeconds: 57.78,
				LongDirection: st.Longitude_WEST,
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := stringToLongitude(tc.inputStr)
			if tc.wantErr {
				if err == nil {
					t.Errorf("Want err from stringToLongitude(%s), got nil", tc.inputStr)
				}
				return
			}
			if err != nil {
				t.Errorf("stringToLongitude(%s) returns unexpected err: %v", tc.inputStr, err)
				return
			}
			if got.Degrees != tc.want.Degrees || got.DegreeMinutes != tc.want.DegreeMinutes || got.DegreeSeconds != tc.want.DegreeSeconds || got.LongDirection != tc.want.LongDirection {
				t.Errorf("Want %v from stringToLongitude(%s), got %v", tc.want, tc.inputStr, got)
			}
		})
	}
}

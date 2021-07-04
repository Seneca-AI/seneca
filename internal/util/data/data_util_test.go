package data

import (
	"fmt"
	st "seneca/api/type"
	"seneca/internal/util"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestCutRawVideoData(t *testing.T) {
	testCases := []struct {
		desc                    string
		rawVideoDuration        time.Duration
		cutVidDuration          time.Duration
		expectedCutVidDurations []time.Duration
	}{
		{
			desc:                    "expected output",
			rawVideoDuration:        ((time.Minute * 5) + (time.Second * 20)),
			cutVidDuration:          time.Minute,
			expectedCutVidDurations: []time.Duration{time.Minute, time.Minute, time.Minute, time.Minute, time.Minute, (time.Second * 20)},
		},
		{
			desc:                    "raw video duration less than cut video duration",
			rawVideoDuration:        time.Second * 20,
			cutVidDuration:          time.Minute,
			expectedCutVidDurations: []time.Duration{time.Second * 20},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			rawVideo := &st.RawVideo{
				UserId:       "user",
				Id:           "rawVidId",
				CreateTimeMs: util.TimeToMilliseconds(time.Date(2021, time.February, 1, 12, 13, 14, 0, time.UTC)),
				DurationMs:   tc.rawVideoDuration.Milliseconds(),
			}

			wantCutVideos := []*st.CutVideo{}
			elapsedTime := time.Duration(0)
			for _, cutVidDur := range tc.expectedCutVidDurations {
				wantCutVideos = append(wantCutVideos, &st.CutVideo{
					UserId:               "user",
					RawVideoId:           "rawVidId",
					CreateTimeMs:         rawVideo.CreateTimeMs + elapsedTime.Milliseconds(),
					DurationMs:           cutVidDur.Milliseconds(),
					CloudStorageFileName: fmt.Sprintf("%s.%d.CUT_VIDEO.mp4", rawVideo.UserId, rawVideo.CreateTimeMs+elapsedTime.Milliseconds()),
				})
				elapsedTime = elapsedTime + cutVidDur
			}

			gotCutVideos := ConstructCutVideoData(tc.cutVidDuration, rawVideo)
			if len(gotCutVideos) != len(wantCutVideos) {
				t.Errorf("Want %d cut videos, got %d", len(wantCutVideos), len(gotCutVideos))
				return
			}
			for i := range gotCutVideos {
				if gotCutVideos[i].RawVideoId != rawVideo.Id || gotCutVideos[i].UserId != wantCutVideos[i].UserId ||
					gotCutVideos[i].CreateTimeMs != wantCutVideos[i].CreateTimeMs || gotCutVideos[i].DurationMs != wantCutVideos[i].DurationMs {
					t.Errorf("Unexpected st.CutVideo (-want, +got) %v", cmp.Diff(wantCutVideos[i], gotCutVideos[i], protocmp.Transform()))
				}
			}
		})
	}
}

func TestLatitudeToFloat(t *testing.T) {
	testCases := []struct {
		desc       string
		lat        *st.Latitude
		wantString string
	}{
		{
			desc: "north just degrees",
			lat: &st.Latitude{
				Degrees:      15,
				LatDirection: st.Latitude_NORTH,
			},
			wantString: "15.000000",
		},
		{
			desc: "north zero",
			lat: &st.Latitude{
				Degrees:      0,
				LatDirection: st.Latitude_NORTH,
			},
			wantString: "0.000000",
		},
		{
			desc: "north degrees and minutes",
			lat: &st.Latitude{
				Degrees:       15,
				DegreeMinutes: 25,
				LatDirection:  st.Latitude_NORTH,
			},
			wantString: "15.416667",
		},
		{
			desc: "north degrees and seconds",
			lat: &st.Latitude{
				Degrees:       45,
				DegreeSeconds: 35,
				LatDirection:  st.Latitude_NORTH,
			},
			wantString: "45.009722",
		},
		{
			desc: "north degrees, minutes and seconds",
			lat: &st.Latitude{
				Degrees:       40,
				DegreeMinutes: 26,
				DegreeSeconds: 13.32,
				LatDirection:  st.Latitude_NORTH,
			},
			wantString: "40.437033",
		},
		{
			desc: "north degrees, minutes and seconds no trailing zeroes",
			lat: &st.Latitude{
				Degrees:       15,
				DegreeMinutes: 6,
				DegreeSeconds: 36,
				LatDirection:  st.Latitude_NORTH,
			},
			wantString: "15.110000",
		},
		{
			desc: "south",
			lat: &st.Latitude{
				Degrees:      15,
				LatDirection: st.Latitude_SOUTH,
			},
			wantString: "-15.000000",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			got := LatitudeToFloat64(tc.lat)
			if fmt.Sprintf("%f", got) != tc.wantString {
				t.Errorf("Want %q, got %f", tc.wantString, got)
			}
		})
	}
}

func TestDistanceMiles(t *testing.T) {
	testCases := []struct {
		desc  string
		lat1  *st.Latitude
		long1 *st.Longitude
		lat2  *st.Latitude
		long2 *st.Longitude
		want  float64
	}{
		{
			desc: "only one test",
			lat1: &st.Latitude{
				Degrees:       50,
				DegreeMinutes: 3,
				DegreeSeconds: 59,
				LatDirection:  st.Latitude_NORTH,
			},
			long1: &st.Longitude{
				Degrees:       5,
				DegreeMinutes: 42,
				DegreeSeconds: 43,
				LongDirection: st.Longitude_WEST,
			},
			lat2: &st.Latitude{
				Degrees:       58,
				DegreeMinutes: 38,
				DegreeSeconds: 38,
			},
			long2: &st.Longitude{
				Degrees:       3,
				DegreeMinutes: 4,
				DegreeSeconds: 12,
				LongDirection: st.Longitude_WEST,
			},
			want: 602,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			got := DistanceMiles(tc.lat1, tc.long1, tc.lat2, tc.long2)
			if got > tc.want+1 || got < tc.want-1 {
				t.Errorf("Want %.45f, got %.45f", tc.want, got)
			}
		})
	}
}

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

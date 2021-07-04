package algorithms

import (
	"fmt"
	"math/rand"
	st "seneca/api/type"
	"seneca/internal/client/intraseneca"
	"seneca/internal/dataprocessor"
	"seneca/internal/util"
	"testing"
	"time"
)

func TestFollowingDistanceV0GenerateDrivingConditions(t *testing.T) {
	if util.IsCIEnv() {
		t.Skip("Skipping exiftool test in GitHub env.")
	}

	userID := "123"
	startTime := time.Date(2021, 06, 05, 0, 0, 0, 0, time.UTC)

	// Generate a few motions and frames with a slight offset to test floor and ceiling buckets.
	rawMotions := []interface{}{}
	rawFrames := []interface{}{}
	for i := 0; i < 100; i++ {
		rawMotion0 := &st.RawMotion{
			UserId: userID,
			Id:     fmt.Sprintf("%d", i),
			Motion: &st.Motion{
				VelocityMph: 47,
			},
			TimestampMs: util.TimeToMilliseconds(startTime.Add(time.Second*time.Duration(i) + time.Duration(rand.Int63n(time.Second.Milliseconds())))),
			Source: &st.Source{
				SourceId:   fmt.Sprintf("%d", i+100),
				SourceType: st.Source_RAW_VIDEO,
			},
		}
		rawMotion1 := &st.RawMotion{
			UserId: userID,
			Id:     fmt.Sprintf("%d", i+200),
			Motion: &st.Motion{
				VelocityMph: 57,
			},
			TimestampMs: util.TimeToMilliseconds(startTime.Add(time.Second*time.Duration(i) + time.Duration(rand.Int63n(time.Second.Milliseconds())))),
			Source: &st.Source{
				SourceId:   fmt.Sprintf("%d", i+300),
				SourceType: st.Source_RAW_VIDEO,
			},
		}
		rawMotions = append(rawMotions, rawMotion0)
		rawMotions = append(rawMotions, rawMotion1)

		rawFrame := &st.RawFrame{
			UserId:      userID,
			Id:          fmt.Sprintf("%d", i+400),
			TimestampMs: util.TimeToMilliseconds(startTime.Add(time.Second*time.Duration(i) + time.Duration(rand.Int63n(time.Second.Milliseconds())))),
			Source: &st.Source{
				SourceId:   fmt.Sprintf("%d", i+500),
				SourceType: st.Source_RAW_VIDEO,
			},
		}

		rawFrames = append(rawFrames, rawFrame)
	}

	allData := map[string][]interface{}{
		dataprocessor.RawMotionTypeString: rawMotions,
		dataprocessor.RawFrameTypeString:  rawFrames,
	}

	mockIntraSeneca := intraseneca.NewMockIntraSenecaClient()
	// Return values that will trigger 'yes'...
	for i := range rawFrames {
		rawFrame := rawFrames[i].(*st.RawFrame)

		request := &st.ObjectsInFrameRequest{
			RawFrame: rawFrame,
		}
		response := &st.ObjectsInFrameResponse{
			ObjectInFrame: &st.ObjectsInFrame{
				ObjectBox: []*st.ObjectBox{
					{
						XLower:      0.42,
						YLower:      0.40,
						XUpper:      0.53,
						YUpper:      0.50,
						ObjectLabel: st.ObjectBox_CAR,
						Confidence:  0.75,
					},
				},
			},
		}

		// ..except for between 30 and 40 and 80 and 90.
		if (i >= 30 && i < 40) || (i >= 80 && i < 90) {
			response.ObjectInFrame.ObjectBox[0].Confidence = 0.25
		}

		mockIntraSeneca.InsertProcessObjectsInFrameResponse(request, response)
	}

	followingDistanceV0, err := newFollowingDistanceV0(mockIntraSeneca)
	if err != nil {
		t.Fatalf("newFollowingDistanceV0() returns err: %v", err)
	}

	drivingConditions, err := followingDistanceV0.GenerateDrivingConditions(allData)
	if err != nil {
		t.Fatalf("newFollowingDistanceV0() returns err: %v", err)
	}

	if len(drivingConditions) != 3 {
		t.Fatalf("Want len(3) drivingConditions, got %d", len(drivingConditions))
	}
}

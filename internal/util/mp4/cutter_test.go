package mp4

import (
	"fmt"
	"seneca/api/constants"
	st "seneca/api/type"
	"seneca/internal/util"
	"testing"
	"time"
)

func TestCutRawVideoProducesExpectedOutputFilePaths(t *testing.T) {
	rawVideo := &st.RawVideo{
		UserId:       "user",
		CreateTimeMs: util.TimeToMilliseconds(time.Date(2021, time.February, 1, 12, 13, 14, 0, time.UTC)),
		DurationMs:   ((time.Minute * 5) + (time.Second * 20)).Milliseconds(),
	}

	_, outputFilePaths, err := CutRawVideo(time.Minute, "rawVideo.mp4", rawVideo, true)
	if err != nil {
		t.Errorf("CutRawVideo(time.Minute, \"rawVideo.mp4\", %v) returns err: %v", rawVideo, err)
	}
	if len(outputFilePaths) != 6 {
		t.Errorf("Want 6 output file paths from CutRawVideo(time.Minute, \"rawVideo.mp4\", %v), got %d", rawVideo, len(outputFilePaths))
	}
	for i, ofp := range outputFilePaths {
		if ofp != fmt.Sprintf("/rawVideo.%d.mp4", i) {
			t.Errorf("Want %q for outputFilePaths, got %q", fmt.Sprintf("/rawVideo.%d.mp4", i), ofp)
		}
	}
}

func TestCutRawVideoRejectsInvalidInput(t *testing.T) {
	var err error
	if _, _, err = CutRawVideo(time.Duration(0), "", nil, true); err == nil {
		t.Errorf("Expected err from CutRawVideo(0, \"\", nil), got nil")
	}

	if _, _, err = CutRawVideo(time.Duration(0), "invalid", nil, true); err == nil {
		t.Errorf("Expected err from CutRawVideo(0, \"invalid\", nil), got nil")
	}

	rawVideo := &st.RawVideo{}
	pathToTestMp4 := "../../../test/testdata/dad_example.mp4"

	if _, _, err = CutRawVideo(time.Duration(0), pathToTestMp4, rawVideo, true); err == nil {
		t.Errorf("Expected err from CutRawVideo(0, %q, %+v) for RawVideo without userID, got nil", pathToTestMp4, rawVideo)
	}

	rawVideo.UserId = "user"
	if _, _, err = CutRawVideo(time.Duration(0), pathToTestMp4, rawVideo, true); err == nil {
		t.Errorf("Expected err from CutRawVideo(0, %q, %+v) for RawVideo with 0 duration, got nil", pathToTestMp4, rawVideo)
	}

	rawVideo.DurationMs = constants.MaxInputVideoDuration.Milliseconds() + 1
	if _, _, err = CutRawVideo(time.Duration(0), pathToTestMp4, rawVideo, true); err == nil {
		t.Errorf("Expected err from CutRawVideo(0, %q, %+v) for RawVideo with too large duration, got nil", pathToTestMp4, rawVideo)
	}

	rawVideo.DurationMs = (time.Minute * 5).Milliseconds()
	rawVideo.CreateTimeMs = util.TimeToMilliseconds(time.Date(2021, time.February, 1, 12, 13, 14, 0, time.UTC))

	if _, _, err = CutRawVideo(time.Minute, "problem space.mp4", rawVideo, true); err == nil {
		t.Errorf("Expected err from CutRawVideo(0, %q, %+v) for RawVideo with space in file name, got nil", pathToTestMp4, rawVideo)
	}
}

package cutter

import (
	"fmt"
	"os"
	"seneca/api/constants"
	st "seneca/api/type"
	"seneca/internal/client/logging"
	"seneca/internal/util"
	"seneca/internal/util/mp4/headerparse"
	"testing"
	"time"
)

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

func TestRawVideoToFrames(t *testing.T) {
	if util.IsCIEnv() {
		t.Skip("Skipping ffmpeg test in GitHub env.")
	}

	userID := "123"
	rawVideoID := "345"

	logger := logging.NewLocalLogger(false)

	pathToVideo := "../../../../test/testdata/blackvue_example.mp4"

	exiftool := headerparse.NewExifMP4Tool(logger)

	rawVideo, _, _, _, err := exiftool.ParseVideoMetadata(pathToVideo)
	if err != nil {
		t.Fatalf(" exiftool.ParseVideoMetadata() returns err: %v", err)
	}
	rawVideo.UserId = userID
	rawVideo.Id = rawVideoID

	rawFrames, tempDirName, fileNames, err := RawVideoToFrames(1, pathToVideo, rawVideo)
	if err != nil {
		t.Fatalf("RawVideoToFrames() returns err: %v", err)
	}

	defer os.RemoveAll(tempDirName)

	if len(rawFrames) != 60 {
		t.Fatalf("Want len(60) for rawFrames, got %d", len(rawFrames))
	}

	for i, rf := range rawFrames {
		if rf.UserId != userID {
			t.Errorf("Want %q for userID, got %q", userID, rf.UserId)
		}

		wantMS := util.TimeToMilliseconds(util.MillisecondsToTime(rawVideo.CreateTimeMs).Add(time.Second * time.Duration(i)))
		if rf.TimestampMs != wantMS {
			t.Errorf("Want timestamp at %v got %v", util.MillisecondsToTime(wantMS), util.MillisecondsToTime(rf.TimestampMs))
		}

		if rf.CloudStorageFileName != fmt.Sprintf("%s.%s.%d.png", userID, rawVideo.Id, rf.TimestampMs) {
			t.Errorf("Want CloudStorageFileName %q, got %q", fmt.Sprintf("%d.%s.png", rf.TimestampMs, userID), rf.CloudStorageFileName)
		}

		if rf.Source.SourceType != st.Source_RAW_VIDEO || rf.Source.SourceId != rawVideoID {
			t.Errorf("Want Source{Id: %s, Type: %s}, got Source{Id: %s, Type: %s", rawVideoID, st.Source_RAW_VIDEO, rf.Source.SourceId, rf.Source.SourceType)
		}
	}

	if len(fileNames) != 60 {
		t.Fatalf("Want %d files in tempDir, got %d", 60, len(fileNames))
	}
}

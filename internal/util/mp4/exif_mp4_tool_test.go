package mp4

import (
	"errors"
	"os"
	"seneca/api/senecaerror"
	"seneca/internal/util"
	"testing"
	"time"
)

func TestGetMetadataHasExpectedData(t *testing.T) {
	if os.Getenv("CI") == "true" {
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
	if os.Getenv("CI") == "true" {
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

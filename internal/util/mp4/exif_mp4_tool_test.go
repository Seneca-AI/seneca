package mp4

import (
	"os"
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
	expectedCreationTime := time.Date(2021, time.February, 13, 17, 47, 49, 0, time.UTC)
	expectedDuration := time.Minute

	videoMetaData, err := exifMP4Tool.GetMetadata(pathToTestMp4)
	if err != nil {
		t.Errorf("GetMetadata(%s) returns err: %v", pathToTestMp4, err)
		return
	}

	if *videoMetaData.CreationTime != expectedCreationTime {
		t.Errorf("videoMetaData.CreationTime incorrect. got %v, want %v", videoMetaData.CreationTime, expectedCreationTime)
	}
	if *videoMetaData.Duration != expectedDuration {
		t.Errorf("videoMetaData.Duration incorrect. got %v, want %v", videoMetaData.Duration, expectedDuration)
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

	if _, err := exifMP4Tool.GetMetadata("../idontexist"); err == nil {
		t.Errorf("Expected non-nil error from bogus input file, got nil")
	}
}

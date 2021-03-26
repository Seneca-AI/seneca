package mp4

import (
	"testing"
	"time"
)

func TestGetMetadataHasExpectedData(t *testing.T) {
	exifMP4Tool, err := NewExitMP4Tool()
	if err != nil {
		t.Errorf("NewExitMP4Tool() returns err: %v", err)
	}

	pathToTestMp4 := "../../../test/testdata/dad_example.MP4"
	expectedCreationTime := time.Date(2021, time.February, 13, 17, 47, 49, 0, time.UTC)
	expectedDuration := time.Minute

	video_meta_data, err := exifMP4Tool.GetMetadata(pathToTestMp4)
	if err != nil {
		t.Errorf("GetMetadata(%s) returns err: %v", pathToTestMp4, err)
		return
	}

	if *video_meta_data.CreationTime != expectedCreationTime {
		t.Errorf("video_meta_data.CreationTime incorrect. got %v, want %v", video_meta_data.CreationTime, expectedCreationTime)
	}
	if *video_meta_data.Duration != expectedDuration {
		t.Errorf("video_meta_data.Duration incorrect. got %v, want %v", video_meta_data.Duration, expectedDuration)
	}
}

func TestGetMetadataDoesntCrashWitoutVideoFile(t *testing.T) {
	exifMP4Tool, err := NewExitMP4Tool()
	if err != nil {
		t.Errorf("NewExitMP4Tool() returns err: %v", err)
	}

	if _, err := exifMP4Tool.GetMetadata("../idontexist"); err == nil {
		t.Errorf("Expected non-nil error from bogus input file, got nil")
	}
}

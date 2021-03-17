package rawvideohandler

import (
	"testing"
	"time"
)

func TestGetMetadataHasExpectedData(t *testing.T) {
	pathToTestMp4 := "./testdata/dad_example.MP4"
	expectedCreationTime := time.Date(2021, time.February, 13, 17, 47, 49, 0, time.UTC)
	expectedDuration := time.Minute

	video_meta_data, err := GetMetadata(pathToTestMp4)
	if err != nil {
		t.Errorf("GetMetadata(%s) returns err: %v", pathToTestMp4, err)
	}

	if video_meta_data.CreationTime != expectedCreationTime {
		t.Errorf("video_meta_data.CreationTime incorrect. got %v, want %v", video_meta_data.CreationTime, expectedCreationTime)
	}
	if video_meta_data.Duration != expectedDuration {
		t.Errorf("video_meta_data.Duration incorrect. got %v, want %v", video_meta_data.Duration, expectedDuration)
	}
}

func TestGetMetadataDoesntCrashWitoutVideoFile(t *testing.T) {
	if _, err := GetMetadata("../idontexist"); err == nil {
		t.Errorf("Expected non-nil error from bogus input file, got nil")
	}
}

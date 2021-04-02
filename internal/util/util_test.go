package util

import (
	"testing"
	"time"
)

func TestGetFileNameFromPath(t *testing.T) {
	testCases := []struct {
		desc    string
		path    string
		want    string
		wantErr bool
	}{
		{
			desc:    "Check valid path",
			path:    "this/is/a/valid/path.mp4",
			want:    "path.mp4",
			wantErr: false,
		},
		{
			desc:    "Check empty path",
			path:    "",
			want:    "",
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := GetFileNameFromPath(tc.path)
			if got != tc.want {
				t.Errorf("want file name %q, got %q", tc.want, got)
			}
			if tc.wantErr == (err == nil) {
				t.Errorf("wantErr (%t), but got %v", tc.wantErr, err)
			}
		})
	}
}

func TestDurationToString(t *testing.T) {
	testCases := []struct {
		dur  time.Duration
		want string
	}{
		{
			dur:  time.Duration(0),
			want: "00:00:00",
		},
		{
			dur:  time.Second * 5,
			want: "00:00:05",
		},
		{
			dur:  ((time.Hour * 5) + (time.Minute * 4) + (time.Second * 3)),
			want: "05:04:03",
		},
		{
			dur:  ((time.Hour * 15) + (time.Minute * 14) + (time.Second * 13)),
			want: "15:14:13",
		},
	}

	for _, tc := range testCases {
		got := DurationToString(tc.dur)
		if got != tc.want {
			t.Errorf("Got %q from DurationToString(%v), want %q", got, tc.dur, tc.want)
		}
	}
}

func TestMillisecondsToDuration(t *testing.T) {
	testCases := []struct {
		durms int64
		want  time.Duration
	}{
		{
			want:  time.Second,
			durms: 1000,
		},
		{
			want:  time.Minute,
			durms: 60000,
		},
		{
			want:  time.Hour,
			durms: 3600000,
		}, {
			want:  (time.Millisecond + time.Second + time.Minute + time.Hour),
			durms: 3661001,
		},
	}
	for _, tc := range testCases {
		got := MillisecondsToDuration(tc.durms)
		if got != tc.want {
			t.Errorf("Got %v for DurationToMilliseconds(%d), want %v", got, tc.durms, tc.want)
		}
	}

}

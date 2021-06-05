package util

import (
	"strings"
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
func TestSortStringsAlphaNumerically(t *testing.T) {
	testCases := []struct {
		in   []string
		want []string
	}{
		{
			in:   []string{"a/b/c/d/011.png", "a/b/c/d/001.png", "a/b/c/d/111.png"},
			want: []string{"a/b/c/d/001.png", "a/b/c/d/011.png", "a/b/c/d/111.png"},
		},
	}

	for _, tc := range testCases {
		got, err := SortStringsAlphaNumerically(tc.in, func(s string) string {
			sParts := strings.Split(s, "/")
			lastPart := sParts[len(sParts)-1]
			leadingZeroesTrimmed := strings.TrimLeft(lastPart, "0")
			return strings.TrimSuffix(leadingZeroesTrimmed, ".png")
		})
		if err != nil {
			t.Fatalf("SortStringsAlphaNumerically(%v) returns err: %v", tc.in, err)
		}

		if len(tc.want) != len(got) {
			t.Fatalf("Want len(%d), got len(%d)", len(tc.want), len(got))
		}

		for i := range tc.want {
			if tc.want[i] != got[i] {
				t.Fatalf("Want %q at index %d, got %q", tc.want[i], i, got[i])
			}
		}

	}
}

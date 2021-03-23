package util

import (
	"testing"
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
		got, err := GetFileNameFromPath(tc.path)
		if got != tc.want {
			t.Errorf("want file name %q, got %q", tc.want, got)
		}
		if tc.wantErr == (err == nil) {
			t.Errorf("wantErr (%t), but got %v", tc.wantErr, err)
		}
	}
}

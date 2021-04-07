package mp4

import (
	"seneca/api/types"
	"testing"
)

func TestStringToLatitude(t *testing.T) {
	testCases := []struct {
		desc     string
		inputStr string
		want     *types.Latitude
		wantErr  bool
	}{
		{
			desc:     "test empty string throws err",
			inputStr: "",
			want:     nil,
			wantErr:  true,
		},
		{
			desc:     "test malformed string without direction throws err",
			inputStr: "40 deg 24' 57.66\"",
			want:     nil,
			wantErr:  true,
		},
		{
			desc:     "test malformed string with extra info throws err",
			inputStr: "40 deg 24' 57.66\" N P",
			want:     nil,
			wantErr:  true,
		},
		{
			desc:     "test malformed string with bad degrees throws err",
			inputStr: "40 degrees 24' 57.66\" N",
			want:     nil,
			wantErr:  true,
		},
		{
			desc:     "test malformed string without degrees symbol throws err",
			inputStr: "40 deg 24 57.66 N",
			want:     nil,
			wantErr:  true,
		},
		{
			desc:     "test string without decimals succeeds",
			inputStr: "40 deg 24' 57\" N",
			want: &types.Latitude{
				Degrees:       40,
				DegreeMinutes: 24,
				DegreeSeconds: 57,
				LatDirection:  types.Latitude_NORTH,
			},
			wantErr: false,
		},
		{
			desc:     "test string with all decimals and south succeeds",
			inputStr: "40.35 deg 24.56' 57.78\" S",
			want: &types.Latitude{
				Degrees:       40.35,
				DegreeMinutes: 24.56,
				DegreeSeconds: 57.78,
				LatDirection:  types.Latitude_SOUTH,
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := StringToLatitude(tc.inputStr)
			if tc.wantErr {
				if err == nil {
					t.Errorf("Want err from StringToLatitude(%s), got nil", tc.inputStr)
				}
				return
			}
			if err != nil {
				t.Errorf("StringToLatitude(%s) returns unexpected err: %v", tc.inputStr, err)
				return
			}
			if got.Degrees != tc.want.Degrees || got.DegreeMinutes != tc.want.DegreeMinutes || got.DegreeSeconds != tc.want.DegreeSeconds || got.LatDirection != tc.want.LatDirection {
				t.Errorf("Want %v from StringToLatitude(%s), got %v", tc.want, tc.inputStr, got)
			}
		})
	}
}

func TestStringToLongitude(t *testing.T) {
	testCases := []struct {
		desc     string
		inputStr string
		want     *types.Longitude
		wantErr  bool
	}{
		{
			desc:     "test empty string throws err",
			inputStr: "",
			want:     nil,
			wantErr:  true,
		},
		{
			desc:     "test malformed string without direction throws err",
			inputStr: "40 deg 24' 57.66\"",
			want:     nil,
			wantErr:  true,
		},
		{
			desc:     "test malformed string with extra info throws err",
			inputStr: "40 deg 24' 57.66\" E P",
			want:     nil,
			wantErr:  true,
		},
		{
			desc:     "test malformed string with bad degrees throws err",
			inputStr: "40 degrees 24' 57.66\" E",
			want:     nil,
			wantErr:  true,
		},
		{
			desc:     "test malformed string without degrees symbol throws err",
			inputStr: "40 deg 24 57.66 E",
			want:     nil,
			wantErr:  true,
		},
		{
			desc:     "test string without decimals succeeds",
			inputStr: "40 deg 24' 57\" E",
			want: &types.Longitude{
				Degrees:       40,
				DegreeMinutes: 24,
				DegreeSeconds: 57,
				LongDirection: types.Longitude_EAST,
			},
			wantErr: false,
		},
		{
			desc:     "test string with all decimals and west succeeds",
			inputStr: "40.35 deg 24.56' 57.78\" W",
			want: &types.Longitude{
				Degrees:       40.35,
				DegreeMinutes: 24.56,
				DegreeSeconds: 57.78,
				LongDirection: types.Longitude_WEST,
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := StringToLongitude(tc.inputStr)
			if tc.wantErr {
				if err == nil {
					t.Errorf("Want err from StringToLongitude(%s), got nil", tc.inputStr)
				}
				return
			}
			if err != nil {
				t.Errorf("StringToLongitude(%s) returns unexpected err: %v", tc.inputStr, err)
				return
			}
			if got.Degrees != tc.want.Degrees || got.DegreeMinutes != tc.want.DegreeMinutes || got.DegreeSeconds != tc.want.DegreeSeconds || got.LongDirection != tc.want.LongDirection {
				t.Errorf("Want %v from StringToLongitude(%s), got %v", tc.want, tc.inputStr, got)
			}
		})
	}
}

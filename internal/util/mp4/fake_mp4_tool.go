package mp4

import (
	"fmt"
	"seneca/api/senecaerror"
	"seneca/api/types"
	"time"
)

type locationsMotionsTimes struct {
	locations []*types.Location
	motions   []*types.Motion
	times     []time.Time
}

type FakeMP4Tool struct {
	// key: path
	rawVideosMap map[string]*types.RawVideo
	// key: path
	locationsMotionsTimesMap map[string]locationsMotionsTimes
}

func NewFakeMP4Tool() *FakeMP4Tool {
	return &FakeMP4Tool{
		rawVideosMap:             make(map[string]*types.RawVideo),
		locationsMotionsTimesMap: make(map[string]locationsMotionsTimes),
	}
}

// ParseOutRawVideoMetadata returns the stored *types.RawVideo from the rawVideosMap for the given path key.
func (fakeMP4Tool *FakeMP4Tool) ParseOutRawVideoMetadata(pathToVideo string) (*types.RawVideo, error) {
	rawVideo, ok := fakeMP4Tool.rawVideosMap[pathToVideo]
	if !ok {
		return nil, senecaerror.NewBadStateError(fmt.Errorf("file at path %q does not exist", pathToVideo))
	}
	return rawVideo, nil
}

// InsertMetadata stores the given *types.RawVideo in the rawVideosMap with key at the given pathToVideo.
func (fakeMP4Tool *FakeMP4Tool) InsertRawVideoMetadata(pathToVideo string, rawVideo *types.RawVideo) {
	fakeMP4Tool.rawVideosMap[pathToVideo] = rawVideo
}

// 	ParseOutGPSMetadata extracts a list of types.Location, types.Motion and time.Time from the video at the given path.
func (fakeMP4Tool *FakeMP4Tool) ParseOutGPSMetadata(pathToVideo string) ([]*types.Location, []*types.Motion, []time.Time, error) {
	lmts, ok := fakeMP4Tool.locationsMotionsTimesMap[pathToVideo]
	if !ok {
		return nil, nil, nil, senecaerror.NewBadStateError(fmt.Errorf("file at path %q does not exist", pathToVideo))
	}
	return lmts.locations, lmts.motions, lmts.times, nil
}

//	InsertGPSMetadata stores the given []*types.Location, []*types.Motion, []time.Time in the locationsMotionsTimesMap with key at the given pathToVideo.
func (fakeMP4Tool *FakeMP4Tool) InsertGPSMetadata(pathToVideo string, locations []*types.Location, motions []*types.Motion, times []time.Time) {
	fakeMP4Tool.locationsMotionsTimesMap[pathToVideo] = locationsMotionsTimes{
		locations: locations,
		motions:   motions,
		times:     times,
	}
}

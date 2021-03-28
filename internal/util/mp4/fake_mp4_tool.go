package mp4

import (
	"fmt"
	"seneca/api/senecaerror"
	"seneca/api/types"
)

type FakeMP4Tool struct {
	// (key: path, value: *types.RawVideo for video at that path)
	rawVideosMap map[string]*types.RawVideo
}

func NewFakeMP4Tool() *FakeMP4Tool {
	return &FakeMP4Tool{
		rawVideosMap: make(map[string]*types.RawVideo),
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

// InsertMetadata stores the given *types.RawVideo in the rawVideosMap with
// key at the given pathToVideo.
func (fakeMP4Tool *FakeMP4Tool) InsertRawVideoMetadata(pathToVideo string, rawVideo *types.RawVideo) {
	fakeMP4Tool.rawVideosMap[pathToVideo] = rawVideo
}

package mp4

import (
	"fmt"
	"seneca/api/types"
	"time"
)

type FakeMP4Tool struct {
	ParseOutRawVideoMetadataMock func(pathToVideo string) (*types.RawVideo, error)
	ParseOutGPSMetadataMock      func(pathToVideo string) ([]*types.Location, []*types.Motion, []time.Time, error)
	CutRawVideoMock              func(cutVideoDur time.Duration, pathToRawVideo string, rawVideo *types.RawVideo) ([]*types.CutVideo, []string, error)
}

func NewFakeMP4Tool() *FakeMP4Tool {
	return &FakeMP4Tool{}
}

func (fakeMP4Tool *FakeMP4Tool) ParseOutRawVideoMetadata(pathToVideo string) (*types.RawVideo, error) {
	if fakeMP4Tool.ParseOutRawVideoMetadataMock == nil {
		return nil, fmt.Errorf("ParseOutRawVideoMetadataMock not set")
	}
	return fakeMP4Tool.ParseOutRawVideoMetadataMock(pathToVideo)
}

func (fakeMP4Tool *FakeMP4Tool) ParseOutGPSMetadata(pathToVideo string) ([]*types.Location, []*types.Motion, []time.Time, error) {
	if fakeMP4Tool.ParseOutGPSMetadataMock == nil {
		return nil, nil, nil, fmt.Errorf("ParseOutGPSMetadataMock not set")
	}
	return fakeMP4Tool.ParseOutGPSMetadataMock(pathToVideo)
}

func (fakeMP4Tool *FakeMP4Tool) CutRawVideo(cutVideoDur time.Duration, pathToRawVideo string, rawVideo *types.RawVideo) ([]*types.CutVideo, []string, error) {
	if fakeMP4Tool.CutRawVideoMock == nil {
		return nil, nil, fmt.Errorf("CutRawVideoMock not set")
	}
	return fakeMP4Tool.CutRawVideoMock(cutVideoDur, pathToRawVideo, rawVideo)
}

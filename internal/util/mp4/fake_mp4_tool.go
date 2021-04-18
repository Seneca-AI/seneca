package mp4

import (
	"fmt"
	st "seneca/api/type"
	"time"
)

type FakeMP4Tool struct {
	ParseOutRawVideoMetadataMock func(pathToVideo string) (*st.RawVideo, error)
	ParseOutGPSMetadataMock      func(pathToVideo string) ([]*st.Location, []*st.Motion, []time.Time, error)
	CutRawVideoMock              func(cutVideoDur time.Duration, pathToRawVideo string, rawVideo *st.RawVideo) ([]*st.CutVideo, []string, error)
}

func NewFakeMP4Tool() *FakeMP4Tool {
	return &FakeMP4Tool{}
}

func (fakeMP4Tool *FakeMP4Tool) ParseOutRawVideoMetadata(pathToVideo string) (*st.RawVideo, error) {
	if fakeMP4Tool.ParseOutRawVideoMetadataMock == nil {
		return nil, fmt.Errorf("ParseOutRawVideoMetadataMock not set")
	}
	return fakeMP4Tool.ParseOutRawVideoMetadataMock(pathToVideo)
}

func (fakeMP4Tool *FakeMP4Tool) ParseOutGPSMetadata(pathToVideo string) ([]*st.Location, []*st.Motion, []time.Time, error) {
	if fakeMP4Tool.ParseOutGPSMetadataMock == nil {
		return nil, nil, nil, fmt.Errorf("ParseOutGPSMetadataMock not set")
	}
	return fakeMP4Tool.ParseOutGPSMetadataMock(pathToVideo)
}

func (fakeMP4Tool *FakeMP4Tool) CutRawVideo(cutVideoDur time.Duration, pathToRawVideo string, rawVideo *st.RawVideo) ([]*st.CutVideo, []string, error) {
	if fakeMP4Tool.CutRawVideoMock == nil {
		return nil, nil, fmt.Errorf("CutRawVideoMock not set")
	}
	return fakeMP4Tool.CutRawVideoMock(cutVideoDur, pathToRawVideo, rawVideo)
}

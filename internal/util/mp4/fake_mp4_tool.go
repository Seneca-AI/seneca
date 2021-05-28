package mp4

import (
	"fmt"
	st "seneca/api/type"
	"time"
)

type FakeMP4Tool struct {
	ParseVideoMetadataMock func(pathToVideo string) (*st.RawVideo, []*st.Location, []*st.Motion, []time.Time, error)
	CutRawVideoMock        func(cutVideoDur time.Duration, pathToRawVideo string, rawVideo *st.RawVideo) ([]*st.CutVideo, []string, error)
}

func NewFakeMP4Tool() *FakeMP4Tool {
	return &FakeMP4Tool{}
}

func (fakeMP4Tool *FakeMP4Tool) ParseVideoMetadata(pathToVideo string) (*st.RawVideo, []*st.Location, []*st.Motion, []time.Time, error) {
	if fakeMP4Tool.ParseVideoMetadataMock == nil {
		return nil, nil, nil, nil, fmt.Errorf("ParseVideoMetadataMock not set")
	}
	return fakeMP4Tool.ParseVideoMetadataMock(pathToVideo)
}

func (fakeMP4Tool *FakeMP4Tool) CutRawVideo(cutVideoDur time.Duration, pathToRawVideo string, rawVideo *st.RawVideo) ([]*st.CutVideo, []string, error) {
	if fakeMP4Tool.CutRawVideoMock == nil {
		return nil, nil, fmt.Errorf("CutRawVideoMock not set")
	}
	return fakeMP4Tool.CutRawVideoMock(cutVideoDur, pathToRawVideo, rawVideo)
}

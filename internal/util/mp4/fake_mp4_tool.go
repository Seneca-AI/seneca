package mp4

import (
	"fmt"
	"seneca/api/senecaerror"
)

type FakeMP4Tool struct {
	// (key: path, value: VideoMetadata for video at that path)
	fileMetadataMap map[string]*VideoMetadata
}

func NewFakeMP4Tool() *FakeMP4Tool {
	return &FakeMP4Tool{
		fileMetadataMap: make(map[string]*VideoMetadata),
	}
}

// GetMetadata returns the stored VideoMetadata from the fileMetadataMap.
func (fakeMP4Tool *FakeMP4Tool) GetMetadata(pathToVideo string) (*VideoMetadata, error) {
	fileMetadata, ok := fakeMP4Tool.fileMetadataMap[pathToVideo]
	if !ok {
		return nil, senecaerror.NewBadStateError(fmt.Errorf("file at path %q does not exist", pathToVideo))
	}
	return fileMetadata, nil
}

// InsertMetadata stores the given VideoMetadata in the fileMetadataMap with
// key at the given pathToVideo.
func (fakeMP4Tool *FakeMP4Tool) InsertMetadata(pathToVideo string, metadata *VideoMetadata) {
	fakeMP4Tool.fileMetadataMap[pathToVideo] = metadata
}

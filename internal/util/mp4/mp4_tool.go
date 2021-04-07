package mp4

import (
	"fmt"
	"seneca/api/types"
	"time"
)

// MP4ToolInterface defines the interface for interacting with MP4 files
// throughout Seneca.
//nolint
type MP4ToolInterface interface {
	// ParseOutRawVideoMetadata extracts *types.RawVideo metadata from the mp4 at the given path.
	// Params:
	// 		string pathToVideo: path to mp4 to get RawVideo metadata from
	// Returns:
	// 		*types.RawVideo: the RawVideo object
	//		error
	ParseOutRawVideoMetadata(pathToVideo string) (*types.RawVideo, error)
	// 	ParseOutGPSMetadata extracts a list of types.Location, types.Motion and time.Time from the video at the given path.
	//	Params:
	//		 pathToVideo string: the video to analyze
	//	Returns:
	//		[]*types.Location
	//		[]*types.Motion
	//		[]*time.Time
	//		error
	ParseOutGPSMetadata(pathToVideo string) ([]*types.Location, []*types.Motion, []time.Time, error)
	// 	CutRawVideo cuts the raw video mp4 into smaller clips.
	// 	Params:
	//		cutVideoDur time.Duration: the duration the cut videos should be
	//		pathToRawVideo string: the path to the mp4 to cut up
	//		rawVideo *types.RawVideo: the RawVideo data to reference
	//	Returns:
	//		[]*types.CutVideo: the cut video data
	//		[]string: the paths to the cut video temp files
	//		error
	CutRawVideo(cutVideoDur time.Duration, pathToRawVideo string, rawVideo *types.RawVideo) ([]*types.CutVideo, []string, error)
}

//nolint
type MP4Tool struct {
	exifTool *ExifMP4Tool
}

func NewMP4Tool() (*MP4Tool, error) {
	et, err := NewExifMP4Tool()
	if err != nil {
		return nil, fmt.Errorf("error initializing ExifMP4Tool, err: %w", err)
	}
	return &MP4Tool{
		exifTool: et,
	}, nil
}

func (mt *MP4Tool) ParseOutRawVideoMetadata(pathToVideo string) (*types.RawVideo, error) {
	return mt.exifTool.ParseOutRawVideoMetadata(pathToVideo)
}

func (mt *MP4Tool) ParseOutGPSMetadata(pathToVideo string) ([]*types.Location, []*types.Motion, []time.Time, error) {
	return mt.exifTool.ParseOutGPSMetadata(pathToVideo)
}

func (mt *MP4Tool) CutRawVideo(cutVideoDur time.Duration, pathToRawVideo string, rawVideo *types.RawVideo) ([]*types.CutVideo, []string, error) {
	return CutRawVideo(cutVideoDur, pathToRawVideo, rawVideo, false)
}

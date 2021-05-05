package mp4

import (
	"fmt"
	st "seneca/api/type"
	"seneca/internal/util/mp4/cutter"
	"seneca/internal/util/mp4/headerparse"
	"time"
)

// MP4ToolInterface defines the interface for interacting with MP4 files
// throughout Seneca.
//nolint
type MP4ToolInterface interface {
	// ParseOutRawVideoMetadata extracts *st.RawVideo metadata from the mp4 at the given path.
	// Params:
	// 		string pathToVideo: path to mp4 to get RawVideo metadata from
	// Returns:
	// 		*st.RawVideo: the RawVideo object
	//		error
	ParseOutRawVideoMetadata(pathToVideo string) (*st.RawVideo, error)
	// 	ParseOutGPSMetadata extracts a list of st.Location, st.Motion and time.Time from the video at the given path.
	//	Params:
	//		 pathToVideo string: the video to analyze
	//	Returns:
	//		[]*st.Location
	//		[]*st.Motion
	//		[]*time.Time
	//		error
	ParseOutGPSMetadata(pathToVideo string) ([]*st.Location, []*st.Motion, []time.Time, error)
	// 	CutRawVideo cuts the raw video mp4 into smaller clips.
	// 	Params:
	//		cutVideoDur time.Duration: the duration the cut videos should be
	//		pathToRawVideo string: the path to the mp4 to cut up
	//		rawVideo *st.RawVideo: the RawVideo data to reference
	//	Returns:
	//		[]*st.CutVideo: the cut video data
	//		[]string: the paths to the cut video temp files
	//		error
	CutRawVideo(cutVideoDur time.Duration, pathToRawVideo string, rawVideo *st.RawVideo) ([]*st.CutVideo, []string, error)
}

//nolint
type MP4Tool struct {
	exifTool *headerparse.ExifMP4Tool
}

func NewMP4Tool() (*MP4Tool, error) {
	et, err := headerparse.NewExifMP4Tool()
	if err != nil {
		return nil, fmt.Errorf("error initializing ExifMP4Tool, err: %w", err)
	}
	return &MP4Tool{
		exifTool: et,
	}, nil
}

func (mt *MP4Tool) ParseOutRawVideoMetadata(pathToVideo string) (*st.RawVideo, error) {
	return mt.exifTool.ParseOutRawVideoMetadata(pathToVideo)
}

func (mt *MP4Tool) ParseOutGPSMetadata(pathToVideo string) ([]*st.Location, []*st.Motion, []time.Time, error) {
	return mt.exifTool.ParseOutGPSMetadata(pathToVideo)
}

func (mt *MP4Tool) CutRawVideo(cutVideoDur time.Duration, pathToRawVideo string, rawVideo *st.RawVideo) ([]*st.CutVideo, []string, error) {
	return cutter.CutRawVideo(cutVideoDur, pathToRawVideo, rawVideo, false)
}

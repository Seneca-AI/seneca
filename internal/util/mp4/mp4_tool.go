package mp4

import (
	st "seneca/api/type"
	"seneca/internal/client/logging"
	"seneca/internal/util/mp4/cutter"
	"seneca/internal/util/mp4/headerparse"
	"time"
)

// MP4ToolInterface defines the interface for interacting with MP4 files
// throughout Seneca.
//nolint
type MP4ToolInterface interface {
	// ParseVideoMetadata extracts metadata from the mp4 at the given path.
	// Params:
	// 		string pathToVideo: path to mp4 to get RawVideo metadata from
	// Returns:
	// 		*st.RawVideo
	// 		[]*st.Location
	//		[]*st.Motion
	//		[]time.Time
	//		error, senecaerror.DevError
	ParseVideoMetadata(pathToVideo string) (*st.RawVideo, []*st.Location, []*st.Motion, []time.Time, error)
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

func NewMP4Tool(logger logging.LoggingInterface) (*MP4Tool, error) {
	et := headerparse.NewExifMP4Tool(logger)
	return &MP4Tool{
		exifTool: et,
	}, nil
}

func (mt *MP4Tool) ParseVideoMetadata(pathToVideo string) (*st.RawVideo, []*st.Location, []*st.Motion, []time.Time, error) {
	return mt.exifTool.ParseVideoMetadata(pathToVideo)
}

func (mt *MP4Tool) CutRawVideo(cutVideoDur time.Duration, pathToRawVideo string, rawVideo *st.RawVideo) ([]*st.CutVideo, []string, error) {
	return cutter.CutRawVideo(cutVideoDur, pathToRawVideo, rawVideo, false)
}

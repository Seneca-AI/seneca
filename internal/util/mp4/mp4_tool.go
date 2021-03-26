package mp4

import (
	"encoding/json"
	"fmt"
	"time"
)

// MP4ToolInterface defines the interface for interacting with MP4 files
// throughout Seneca.
type MP4ToolInterface interface {
	// GetMetadata extracts VideoMetadata from the video at the given path.
	// Params:
	// 		string pathToVideo: path to video to get metadata from
	// Returns:
	// 		*VideoMetadata: the VideoMetadata object
	//		error
	GetMetadata(path string) (*VideoMetadata, error)
}

// VideoMetadata holds relevant metadata for mp4 videos.
type VideoMetadata struct {
	CreationTime *time.Time
	Duration     *time.Duration
	Locations    []location
	SpeedsMPH    []int64
}

type location struct {
	lat       string
	long      string
	timestamp time.Time
}

func (vmd *VideoMetadata) String() string {
	// TODO: handle this error with a logger
	b, _ := json.MarshalIndent(vmd, "", "\t")
	return fmt.Sprint(string(b))
}

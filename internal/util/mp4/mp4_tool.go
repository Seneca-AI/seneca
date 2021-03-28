package mp4

import "seneca/api/types"

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
}

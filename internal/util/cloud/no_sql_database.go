package cloud

import (
	"time"

	"seneca/api/types"
)

// NoSQLDatabaseInterface is the interface used for interacting with
// NoSQL Databases across Seneca.
type NoSQLDatabaseInterface interface {
	// GetRawVideo gets the *types.RawVideo for the given user around the specified createTime.
	// Params:
	//		userID string: the userID associated with this video
	//		createTime time.Time: the approximate time the video was created
	// Returns:
	//		*types.RawVideo: the raw video object
	//		error  [ senecaerror.NotFoundError ]
	GetRawVideo(userID string, createTime time.Time) (*types.RawVideo, error)

	// DeleteRawVideoByID deletes the rawVideoo with the given ID from the datastore.
	// Params:
	//		id string
	// Returns:
	//		error
	DeleteRawVideoByID(id string) error

	// GetRawVideoByID gets the rawVideo with the given ID from the datastore.
	// Params:
	//		id string
	// Returns:
	//		*types.RawVideo
	//		error
	GetRawVideoByID(id string) (*types.RawVideo, error)

	// InsertRawVideo inserts the given *types.RawVideo into the RawVideos Directory.
	// Params:
	// 		rawVideo *types.RawVideo: the rawVideo
	// Returns:
	//		string: the newly generated datastore ID for the rawVideo
	//		error
	InsertRawVideo(rawVideo *types.RawVideo) (string, error)

	// InsertUniqueRawVideo inserts the given *types.RawVideo into the RawVideos Directory if a
	// similar RawVideo doesn't already exist.
	// Params:
	// 		rawVideo *types.RawVideo: the rawVideo
	// Returns:
	//		string: the newly generated datastore ID for the rawVideo
	//		error
	InsertUniqueRawVideo(rawVideo *types.RawVideo) (string, error)

	// GetCutVideo gets the *types.CutVideo for the given user around the specified createTime.
	// Params:
	//		userID string: the userID associated with this video
	//		createTime time.Time: the approximate time the video was created
	// Returns:
	//		*types.CutVideo
	//		error: [ senecaerror.NotFoundError ]
	GetCutVideo(userID string, createTime time.Time) (*types.CutVideo, error)

	// DeleteCutVideoByID deletes the cut video with the given ID from the datastore.
	// Params:
	//		id string
	// Returns:
	//		error
	DeleteCutVideoByID(id string) error

	// InsertCutVideo inserts the given *types.CutVideo into the CutVideos directory of the datastore.
	// Params:
	// 		cutVideo *types.CutVideo
	// Returns:
	//		string: the newly generated datastore ID for the cutVideo
	//		error
	InsertCutVideo(cutVideo *types.CutVideo) (string, error)

	// InsertUniqueCutVideo inserts the given *types.CutVideo if a CutVideo with a similar creation time doesn't already exist.
	// Params:
	// 		rawVideo *types.RawVideo
	// Returns:
	//		string: the newly generated datastore ID for the rawVideo
	//		error
	InsertUniqueCutVideo(cutVideo *types.CutVideo) (string, error)

	// GetRawMotion gets the *types.RawMotion for the given user at the given timestamp.
	// Params:
	//		userID string: the userID associated with this RawMotion
	//		timestamp time.Time: the exact time the RawMotion took place
	// Returns:
	//		*types.RawMotion
	//		error: [ senecaerror.NotFoundError ]
	GetRawMotion(userID string, timestamp time.Time) (*types.RawMotion, error)

	// DeleteRawMotionByID deletes the raw motion with the given ID.
	// Params:
	//		id string
	// Returns:
	//		error
	DeleteRawMotionByID(id string) error

	// InsertRawMotion inserts the given *types.RawMotion.
	// Params:
	// 		rawMotion *types.RawMotion
	// Returns:
	//		string: the newly generated datastore ID for the raw motion
	//		error
	InsertRawMotion(rawMotion *types.RawMotion) (string, error)

	// InsertUniqueRawMotion inserts the given *types.RawMotion if a RawMotion with the same creation time doesn't already exist.
	// Params:
	// 		rawMotion *types.RawMotion
	// Returns:
	//		string: the newly generated datastore ID for the raw motion
	//		error
	InsertUniqueRawMotion(rawMotion *types.RawMotion) (string, error)

	// GetRawLocation gets the *types.RawLocation for the given user at the specified timestamp.
	// Params:
	//		userID string: the userID associated with this RawLocation
	//		timestamp time.Time: the exact time the RawLocation took place
	// Returns:
	//		*types.RawLocation
	//		error: [ senecaerror.NotFoundError ]
	GetRawLocation(userID string, createTime time.Time) (*types.RawLocation, error)

	// DeleteRawLocationByID deletes the raw location with the given ID from the datastore.
	// Params:
	//		id string
	// Returns:
	//		error
	DeleteRawLocationByID(id string) error

	// InsertRawLocation inserts the given *types.RawLocation into the RawLocations directory.
	// Params:
	// 		rawLocation *types.RawLocation
	// Returns:
	//		string: the newly generated datastore ID for the raw location
	//		error
	InsertRawLocation(rawLocation *types.RawLocation) (string, error)

	// InsertUniqueRawLocation inserts the given *types.RawLocation if a RawLocation with the same creation time doesn't already exist.
	// Params:
	// 		rawLocation *types.RawLocation
	// Returns:
	//		string: the newly generated datastore ID for the raw location
	//		error
	InsertUniqueRawLocation(rawLocation *types.RawLocation) (string, error)
}

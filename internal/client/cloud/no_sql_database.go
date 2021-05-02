package cloud

import (
	st "seneca/api/type"
	"time"
)

// NoSQLDatabaseInterface is the interface used for interacting with
// NoSQL Databases across Seneca.
type NoSQLDatabaseInterface interface {
	// GetRawVideo gets the *st.RawVideo for the given user around the specified createTime.
	// Params:
	//		userID string: the userID associated with this video
	//		createTime time.Time: the approximate time the video was created
	// Returns:
	//		*st.RawVideo: the raw video object
	//		error  [ senecaerror.NotFoundError ]
	GetRawVideo(userID string, createTime time.Time) (*st.RawVideo, error)

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
	//		*st.RawVideo
	//		error
	GetRawVideoByID(id string) (*st.RawVideo, error)

	// InsertRawVideo inserts the given *st.RawVideo into the RawVideos Directory.
	// Params:
	// 		rawVideo *st.RawVideo: the rawVideo
	// Returns:
	//		string: the newly generated datastore ID for the rawVideo
	//		error
	InsertRawVideo(rawVideo *st.RawVideo) (string, error)

	// InsertUniqueRawVideo inserts the given *st.RawVideo into the RawVideos Directory if a
	// similar RawVideo doesn't already exist.
	// Params:
	// 		rawVideo *st.RawVideo: the rawVideo
	// Returns:
	//		string: the newly generated datastore ID for the rawVideo
	//		error
	InsertUniqueRawVideo(rawVideo *st.RawVideo) (string, error)

	// GetCutVideo gets the *st.CutVideo for the given user around the specified createTime.
	// Params:
	//		userID string: the userID associated with this video
	//		createTime time.Time: the approximate time the video was created
	// Returns:
	//		*st.CutVideo
	//		error: [ senecaerror.NotFoundError ]
	GetCutVideo(userID string, createTime time.Time) (*st.CutVideo, error)

	// DeleteCutVideoByID deletes the cut video with the given ID from the datastore.
	// Params:
	//		id string
	// Returns:
	//		error
	DeleteCutVideoByID(id string) error

	// InsertCutVideo inserts the given *st.CutVideo into the CutVideos directory of the datastore.
	// Params:
	// 		cutVideo *st.CutVideo
	// Returns:
	//		string: the newly generated datastore ID for the cutVideo
	//		error
	InsertCutVideo(cutVideo *st.CutVideo) (string, error)

	// InsertUniqueCutVideo inserts the given *st.CutVideo if a CutVideo with a similar creation time doesn't already exist.
	// Params:
	// 		rawVideo *st.RawVideo
	// Returns:
	//		string: the newly generated datastore ID for the rawVideo
	//		error
	InsertUniqueCutVideo(cutVideo *st.CutVideo) (string, error)

	// GetRawMotion gets the *st.RawMotion for the given user at the given timestamp.
	// Params:
	//		userID string: the userID associated with this RawMotion
	//		timestamp time.Time: the exact time the RawMotion took place
	// Returns:
	//		*st.RawMotion
	//		error: [ senecaerror.NotFoundError ]
	GetRawMotion(userID string, timestamp time.Time) (*st.RawMotion, error)

	// DeleteRawMotionByID deletes the raw motion with the given ID.
	// Params:
	//		id string
	// Returns:
	//		error
	DeleteRawMotionByID(id string) error

	// InsertRawMotion inserts the given *st.RawMotion.
	// Params:
	// 		rawMotion *st.RawMotion
	// Returns:
	//		string: the newly generated datastore ID for the raw motion
	//		error
	InsertRawMotion(rawMotion *st.RawMotion) (string, error)

	// InsertUniqueRawMotion inserts the given *st.RawMotion if a RawMotion with the same creation time doesn't already exist.
	// Params:
	// 		rawMotion *st.RawMotion
	// Returns:
	//		string: the newly generated datastore ID for the raw motion
	//		error
	InsertUniqueRawMotion(rawMotion *st.RawMotion) (string, error)

	// GetRawLocation gets the *st.RawLocation for the given user at the specified timestamp.
	// Params:
	//		userID string: the userID associated with this RawLocation
	//		timestamp time.Time: the exact time the RawLocation took place
	// Returns:
	//		*st.RawLocation
	//		error: [ senecaerror.NotFoundError ]
	GetRawLocation(userID string, createTime time.Time) (*st.RawLocation, error)

	// DeleteRawLocationByID deletes the raw location with the given ID from the datastore.
	// Params:
	//		id string
	// Returns:
	//		error
	DeleteRawLocationByID(id string) error

	// InsertRawLocation inserts the given *st.RawLocation into the RawLocations directory.
	// Params:
	// 		rawLocation *st.RawLocation
	// Returns:
	//		string: the newly generated datastore ID for the raw location
	//		error
	InsertRawLocation(rawLocation *st.RawLocation) (string, error)

	// InsertUniqueRawLocation inserts the given *st.RawLocation if a RawLocation with the same creation time doesn't already exist.
	// Params:
	// 		rawLocation *st.RawLocation
	// Returns:
	//		string: the newly generated datastore ID for the raw location
	//		error
	InsertUniqueRawLocation(rawLocation *st.RawLocation) (string, error)
}

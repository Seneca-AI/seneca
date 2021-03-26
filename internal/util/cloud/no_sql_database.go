package cloud

import (
	"time"

	"seneca/api/types"
)

// NoSQLDatabaseInterface is the interface used for interacting with
// NoSQL Databases across Seneca.
type NoSQLDatabaseInterface interface {
	InsertRawVideo(rawVideo *types.RawVideo) (string, error)
	GetRawVideo(userID string, createTime time.Time) (*types.RawVideo, error)
	InsertUniqueRawVideo(rawVideo *types.RawVideo) (string, error)
	DeleteRawVideoByID(id string) error
}

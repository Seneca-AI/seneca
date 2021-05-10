// Package dao handles logic for accessing data objects.
package dao

import (
	"seneca/api/constants"
	st "seneca/api/type"
	"seneca/internal/client/cloud"
)

type SQLInterface interface {
	ListIDs(tableName constants.TableName, queryParams []*cloud.QueryParam) ([]string, error)
	GetByID(tableName constants.TableName, id string) (interface{}, error)
	Create(tableName constants.TableName, object interface{}) (string, error)
	Insert(tableName constants.TableName, id string, object interface{}) error
	DeleteByID(tableName constants.TableName, id string) error
}

type UserDAO interface {
	InsertUniqueUser(user *st.User) (*st.User, error)
	GetUserByID(id string) (*st.User, error)
	ListAllUserIDs() ([]string, error)
}

type RawVideoDAO interface {
	InsertUniqueRawVideo(rawVideo *st.RawVideo) (*st.RawVideo, error)
	GetRawVideoByID(id string) (*st.RawVideo, error)
	ListUserRawVideoIDs(userID string) ([]string, error)
	DeleteRawVideoByID(id string) error
}

type RawLocationDAO interface {
	InsertUniqueRawLocation(rawLocation *st.RawLocation) (*st.RawLocation, error)
	GetRawLocationByID(id string) (*st.RawLocation, error)
	ListUserRawLocationIDs(userID string) ([]string, error)
	DeleteRawLocationByID(id string) error
}

type RawMotionDAO interface {
	InsertUniqueRawMotion(rawMotion *st.RawMotion) (*st.RawMotion, error)
	GetRawMotionByID(id string) (*st.RawMotion, error)
	ListUserRawMotionIDs(userID string) ([]string, error)
	DeleteRawMotionByID(id string) error
}

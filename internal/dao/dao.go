// Package dao handles logic for accessing data objects.
package dao

import (
	"context"
	st "seneca/api/type"
	"time"
)

// TODO(lucaloncar): remove parent ID params, but enforce parent IDs upon insertion

type AllDAOSet struct {
	UserDAO             UserDAO
	RawVideoDAO         RawVideoDAO
	RawLocationDAO      RawLocationDAO
	RawFrameDAO         RawFrameDAO
	RawMotionDAO        RawMotionDAO
	TripDAO             TripDAO
	EventDAO            EventDAO
	DrivingConditionDAO DrivingConditionDAO
}

type UserDAO interface {
	InsertUniqueUser(user *st.User) (*st.User, error)
	GetUserByID(id string) (*st.User, error)
	ListAllUserIDs() ([]string, error)
	GetUserByEmail(email string) (*st.User, error)
}

type RawVideoDAO interface {
	InsertUniqueRawVideo(rawVideo *st.RawVideo) (*st.RawVideo, error)
	PutRawVideoByID(ctx context.Context, rawVideoID string, rawVideo *st.RawVideo) error
	GetRawVideoByID(id string) (*st.RawVideo, error)
	ListUnprocessedRawVideoIDs(userID string, latestVersion float64) ([]string, error)
	ListUserRawVideoIDs(userID string) ([]string, error)
	DeleteRawVideoByID(id string) error
}

type RawLocationDAO interface {
	InsertUniqueRawLocation(rawLocation *st.RawLocation) (*st.RawLocation, error)
	PutRawLocationByID(ctx context.Context, rawLocationID string, rawLocation *st.RawLocation) error
	GetRawLocationByID(id string) (*st.RawLocation, error)
	ListUnprocessedRawLocationsIDs(userID string, latestVersion float64) ([]string, error)
	ListUserRawLocationIDs(userID string) ([]string, error)
	DeleteRawLocationByID(id string) error
}

type RawFrameDAO interface {
	InsertUniqueRawFrame(rawFrame *st.RawFrame) (*st.RawFrame, error)
	PutRawFrameByID(ctx context.Context, rawFrameID string, rawFrame *st.RawFrame) error
	GetRawFrameByID(id string) (*st.RawFrame, error)
	ListUnprocessedRawFramesIDs(userID string, latestVersion float64) ([]string, error)
	ListUserRawFrameIDs(userID string) ([]string, error)
	DeleteRawFrameByID(id string) error
}

type RawMotionDAO interface {
	InsertUniqueRawMotion(rawMotion *st.RawMotion) (*st.RawMotion, error)
	PutRawMotionByID(ctx context.Context, rawMotionID string, rawMotion *st.RawMotion) error
	ListUnprocessedRawMotionIDs(userID string, latestVersion float64) ([]string, error)
	GetRawMotionByID(id string) (*st.RawMotion, error)
	ListUserRawMotionIDs(userID string) ([]string, error)
	DeleteRawMotionByID(id string) error
}

type TripDAO interface {
	CreateUniqueTrip(ctx context.Context, trip *st.TripInternal) (*st.TripInternal, error)
	PutTripByID(ctx context.Context, tripID string, trip *st.TripInternal) error
	GetTripByID(userID, tripID string) (*st.TripInternal, error)
	ListUserTripIDs(userID string) ([]string, error)
	ListUserTripIDsByTime(userID string, startTime time.Time, endTime time.Time) ([]string, error)
	DeleteTripByID(ctx context.Context, tripID string) error
}

type EventDAO interface {
	CreateEvent(ctx context.Context, event *st.EventInternal) (*st.EventInternal, error)
	GetEventByID(userID, tripID, eventID string) (*st.EventInternal, error)
	ListTripEventIDs(userID, tripID string) ([]string, error)
	DeleteEventByID(ctx context.Context, userID, tripID, eventID string) error
	PutEventByID(ctx context.Context, userID, tripID, eventID string, event *st.EventInternal) error
}

type DrivingConditionDAO interface {
	CreateDrivingCondition(ctx context.Context, drivingCondition *st.DrivingConditionInternal) (*st.DrivingConditionInternal, error)
	GetDrivingConditionByID(userID, tripID, drivingConditionID string) (*st.DrivingConditionInternal, error)
	ListTripDrivingConditionIDs(userID, tripID string) ([]string, error)
	DeleteDrivingConditionByID(ctx context.Context, userID, tripID, eventID string) error
}

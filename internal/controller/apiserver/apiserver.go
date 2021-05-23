package apiserver

import (
	"fmt"
	st "seneca/api/type"
	"seneca/internal/dao"
	"seneca/internal/dataaggregator/sanitizer"
	"time"
)

type APIServer struct {
	sanitizer *sanitizer.Sanitizer
	tripDAO   dao.TripDAO
}

func New(sanitizer *sanitizer.Sanitizer, tripDAO dao.TripDAO) *APIServer {
	return &APIServer{
		sanitizer: sanitizer,
		tripDAO:   tripDAO,
	}
}

func (srv *APIServer) ListTrips(userID string, startTime time.Time, endTime time.Time) ([]*st.Trip, error) {
	trips := []*st.Trip{}

	tripIDs, err := srv.tripDAO.ListUserTripIDsByTime(userID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("ListUserTripIDsByTime(%s, %s, %s) returns err: %w", userID, startTime, endTime, err)
	}

	for _, tid := range tripIDs {
		tripInternal, err := srv.tripDAO.GetTripByID(userID, tid)
		if err != nil {
			return nil, fmt.Errorf("GetTripByID(%s) returns err: %w", tid, err)
		}
		tripExternal, err := srv.sanitizer.TripInternalToTripExternal(tripInternal)
		if err != nil {
			return nil, fmt.Errorf("error converting internal trip %v to external trip: %w", tripInternal, err)
		}
		trips = append(trips, tripExternal)
	}

	return trips, nil
}

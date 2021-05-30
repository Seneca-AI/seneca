package rawlocationdao

import (
	"context"
	"log"
	st "seneca/api/type"
)

type MockRawLocatinDAO struct {
	InsertUniqueRawLocationMock        func(rawLocation *st.RawLocation) (*st.RawLocation, error)
	PutRawLocationByIDMock             func(ctx context.Context, rawLocationID string, rawLocation *st.RawLocation) error
	GetRawLocationByIDMock             func(id string) (*st.RawLocation, error)
	ListUserRawLocationIDsMock         func(userID string) ([]string, error)
	DeleteRawLocationByIDMock          func(id string) error
	ListUnprocessedRawLocationsIDsMock func(userID string, latestVersion float64) ([]string, error)
}

func (mrld *MockRawLocatinDAO) InsertUniqueRawLocation(rawLocation *st.RawLocation) (*st.RawLocation, error) {
	if mrld.InsertUniqueRawLocationMock == nil {
		log.Fatal("InsertUniqueRawLocationMock called but not set")
	}
	return mrld.InsertUniqueRawLocationMock(rawLocation)
}

func (mrld *MockRawLocatinDAO) GetRawLocationByID(id string) (*st.RawLocation, error) {
	if mrld.GetRawLocationByIDMock == nil {
		log.Fatal("GetRawLocationByIDMock called but not set")
	}
	return mrld.GetRawLocationByIDMock(id)
}

func (mrld *MockRawLocatinDAO) ListUserRawLocationIDs(userID string) ([]string, error) {
	if mrld.ListUserRawLocationIDsMock == nil {
		log.Fatal("ListUserRawLocationIDsMock called but not set")
	}
	return mrld.ListUserRawLocationIDsMock(userID)
}

func (mrld *MockRawLocatinDAO) ListUnprocessedRawLocationsIDs(userID string, latestVersion float64) ([]string, error) {
	if mrld.ListUnprocessedRawLocationsIDsMock == nil {
		log.Fatal("ListUnprocessedRawLocationsIDsMock called but not set")
	}
	return mrld.ListUnprocessedRawLocationsIDsMock(userID, latestVersion)
}

func (mrld *MockRawLocatinDAO) DeleteRawLocationByID(id string) error {
	if mrld.DeleteRawLocationByIDMock == nil {
		log.Fatal("DeleteRawLocationByIDMock called but not set")
	}
	return mrld.DeleteRawLocationByIDMock(id)
}

func (mrld *MockRawLocatinDAO) PutRawLocationByID(ctx context.Context, rawLocationID string, rawLocation *st.RawLocation) error {
	if mrld.PutRawLocationByIDMock == nil {
		log.Fatal("PutRawLocationByIDMock called but not set")
	}
	return mrld.PutRawLocationByIDMock(ctx, rawLocationID, rawLocation)
}

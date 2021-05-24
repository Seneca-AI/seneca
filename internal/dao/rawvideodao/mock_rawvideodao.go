package rawvideodao

import (
	"context"
	"log"
	st "seneca/api/type"
)

type MockRawVideoDAO struct {
	InsertUniqueRawVideoMock       func(rawVideo *st.RawVideo) (*st.RawVideo, error)
	GetRawVideoByIDMock            func(id string) (*st.RawVideo, error)
	ListUserRawVideoIDsMock        func(userID string) ([]string, error)
	DeleteRawVideoByIDMock         func(id string) error
	PutRawVideoByIDMock            func(ctx context.Context, rawVideoID string, rawVideo *st.RawVideo) error
	ListUnprocessedRawVideoIDsMock func(userID string, latestVersion float64) ([]string, error)
}

func (mrvd *MockRawVideoDAO) InsertUniqueRawVideo(rawVideo *st.RawVideo) (*st.RawVideo, error) {
	if mrvd.InsertUniqueRawVideoMock == nil {
		log.Fatal("InsertUniqueRawVideoMock called but not set")
	}
	return mrvd.InsertUniqueRawVideoMock(rawVideo)
}

func (mrvd *MockRawVideoDAO) GetRawVideoByID(id string) (*st.RawVideo, error) {
	if mrvd.GetRawVideoByIDMock == nil {
		log.Fatal("InsertUniqueRawVideoMock called but not set")
	}
	return mrvd.GetRawVideoByIDMock(id)
}

func (mrvd *MockRawVideoDAO) ListUserRawVideoIDs(userID string) ([]string, error) {
	if mrvd.ListUserRawVideoIDsMock == nil {
		log.Fatal("ListUserRawVideoIDsMock called but not set")
	}
	return mrvd.ListUserRawVideoIDsMock(userID)
}

func (mrvd *MockRawVideoDAO) DeleteRawVideoByID(id string) error {
	if mrvd.DeleteRawVideoByIDMock == nil {
		log.Fatal("DeleteRawVideoByIDMock called but not set")
	}
	return mrvd.DeleteRawVideoByIDMock(id)
}

func (mrvd *MockRawVideoDAO) PutRawVideoByID(ctx context.Context, rawVideoID string, rawVideo *st.RawVideo) error {
	if mrvd.PutRawVideoByIDMock == nil {
		log.Fatal("PutRawVideoByIDMock called but not set")
	}
	return mrvd.PutRawVideoByIDMock(ctx, rawVideoID, rawVideo)
}

func (mrvd *MockRawVideoDAO) ListUnprocessedRawVideoIDs(userID string, latestVersion float64) ([]string, error) {
	if mrvd.ListUnprocessedRawVideoIDsMock == nil {
		log.Fatal("ListUnprocessedRawVideoIDsMock called but not set")
	}
	return mrvd.ListUnprocessedRawVideoIDsMock(userID, latestVersion)
}

package rawvideodao

import (
	"log"
	st "seneca/api/type"
)

type MockRawVideoDAO struct {
	InsertUniqueRawVideoMock func(rawVideo *st.RawVideo) (*st.RawVideo, error)
	GetRawVideoByIDMock      func(id string) (*st.RawVideo, error)
	ListUserRawVideoIDsMock  func(userID string) ([]string, error)
	DeleteRawVideoByIDMock   func(id string) error
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

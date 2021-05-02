package cloud

import (
	"fmt"
	st "seneca/api/type"
	"time"
)

//	FakeNoSQLDatabaseClient mocks NoSQLDatabaseInterface.
type FakeNoSQLDatabaseClient struct {
	DeleteRawVideoByIDMock      func(id string) error
	GetRawVideoMock             func(userID string, createTime time.Time) (*st.RawVideo, error)
	GetRawVideoByIDMock         func(id string) (*st.RawVideo, error)
	InsertRawVideoMock          func(rawVideo *st.RawVideo) (string, error)
	InsertUniqueRawVideoMock    func(rawVideo *st.RawVideo) (string, error)
	DeleteCutVideoByIDMock      func(id string) error
	GetCutVideoMock             func(userID string, createTime time.Time) (*st.CutVideo, error)
	InsertCutVideoMock          func(cutVideo *st.CutVideo) (string, error)
	InsertUniqueCutVideoMock    func(cutVideo *st.CutVideo) (string, error)
	DeleteRawLocationByIDMock   func(id string) error
	GetRawLocationMock          func(userID string, timestamp time.Time) (*st.RawLocation, error)
	InsertRawLocationMock       func(rawLocation *st.RawLocation) (string, error)
	InsertUniqueRawLocationMock func(rawLocation *st.RawLocation) (string, error)
	DeleteRawMotionByIDMock     func(id string) error
	GetRawMotionMock            func(userID string, timestamp time.Time) (*st.RawMotion, error)
	InsertRawMotionMock         func(rawMotion *st.RawMotion) (string, error)
	InsertUniqueRawMotionMock   func(rawMotion *st.RawMotion) (string, error)
	ListAllUserIDsMock          func(pageToken string, maxResults int) ([]string, string, error)
	GetUserByIDMock             func(id string) (*st.User, error)
}

//	NewFakeNoSQLDatabaseClient returns an instance of FakeNoSQLDatabaseClient.
//	Params:
//	Returns:
//		 *FakeNoSQLDatabaseClient
func NewFakeNoSQLDatabaseClient() *FakeNoSQLDatabaseClient {
	return &FakeNoSQLDatabaseClient{}
}

func (fnsdc *FakeNoSQLDatabaseClient) DeleteRawVideoByID(id string) error {
	if fnsdc.DeleteRawVideoByIDMock == nil {
		return fmt.Errorf("DeleteRawVideoByIDMock not set")
	}
	return fnsdc.DeleteRawVideoByIDMock(id)
}

func (fnsdc *FakeNoSQLDatabaseClient) GetRawVideo(userID string, createTime time.Time) (*st.RawVideo, error) {
	if fnsdc.GetRawVideoMock == nil {
		return nil, fmt.Errorf("GetRawVideoMock not set")
	}
	return fnsdc.GetRawVideoMock(userID, createTime)
}

func (fnsdc *FakeNoSQLDatabaseClient) GetRawVideoByID(id string) (*st.RawVideo, error) {
	if fnsdc.GetRawVideoByIDMock == nil {
		return nil, fmt.Errorf("GetRawVideoByIDMock not set")
	}
	return fnsdc.GetRawVideoByIDMock(id)
}

func (fnsdc *FakeNoSQLDatabaseClient) InsertRawVideo(rawVideo *st.RawVideo) (string, error) {
	if fnsdc.InsertRawVideoMock == nil {
		return "", fmt.Errorf("InsertRawVideoMock not set")
	}
	return fnsdc.InsertRawVideoMock(rawVideo)
}

func (fnsdc *FakeNoSQLDatabaseClient) InsertUniqueRawVideo(rawVideo *st.RawVideo) (string, error) {
	if fnsdc.InsertUniqueRawVideoMock == nil {
		return "", fmt.Errorf("InsertUniqueRawVideoMock not set")
	}
	return fnsdc.InsertUniqueRawVideoMock(rawVideo)
}

func (fnsdc *FakeNoSQLDatabaseClient) DeleteCutVideoByID(id string) error {
	if fnsdc.DeleteCutVideoByIDMock == nil {
		return fmt.Errorf("DeleteCutVideoByIDMock not set")
	}
	return fnsdc.DeleteCutVideoByIDMock(id)
}

func (fnsdc *FakeNoSQLDatabaseClient) GetCutVideo(userID string, createTime time.Time) (*st.CutVideo, error) {
	if fnsdc.GetCutVideoMock == nil {
		return nil, fmt.Errorf("GetCutVideoMock not set")
	}
	return fnsdc.GetCutVideoMock(userID, createTime)
}

func (fnsdc *FakeNoSQLDatabaseClient) InsertCutVideo(cutVideo *st.CutVideo) (string, error) {
	if fnsdc.InsertCutVideoMock == nil {
		return "", fmt.Errorf("InsertCutVideoMock not set")
	}
	return fnsdc.InsertCutVideoMock(cutVideo)
}

func (fnsdc *FakeNoSQLDatabaseClient) InsertUniqueCutVideo(cutVideo *st.CutVideo) (string, error) {
	if fnsdc.InsertUniqueCutVideoMock == nil {
		return "", fmt.Errorf("InsertUniqueCutVideoMock not set")
	}
	return fnsdc.InsertUniqueCutVideoMock(cutVideo)
}

func (fnsdc *FakeNoSQLDatabaseClient) DeleteRawLocationByID(id string) error {
	if fnsdc.DeleteRawLocationByIDMock == nil {
		return fmt.Errorf("DeleteRawLocationByIDMock not set")
	}
	return fnsdc.DeleteRawLocationByIDMock(id)
}

func (fnsdc *FakeNoSQLDatabaseClient) GetRawLocation(userID string, timestamp time.Time) (*st.RawLocation, error) {
	if fnsdc.GetRawLocationMock == nil {
		return nil, fmt.Errorf("GetRawLocationMock not set")
	}
	return fnsdc.GetRawLocationMock(userID, timestamp)
}

func (fnsdc *FakeNoSQLDatabaseClient) InsertRawLocation(rawLocation *st.RawLocation) (string, error) {
	if fnsdc.InsertRawLocationMock == nil {
		return "", fmt.Errorf("InsertRawLocationMock not set")
	}
	return fnsdc.InsertRawLocationMock(rawLocation)
}

func (fnsdc *FakeNoSQLDatabaseClient) InsertUniqueRawLocation(rawLocation *st.RawLocation) (string, error) {
	if fnsdc.InsertUniqueRawLocationMock == nil {
		return "", fmt.Errorf("InsertUniqueRawLocationMock not set")
	}
	return fnsdc.InsertUniqueRawLocationMock(rawLocation)
}

func (fnsdc *FakeNoSQLDatabaseClient) DeleteRawMotionByID(id string) error {
	if fnsdc.DeleteRawMotionByIDMock == nil {
		return fmt.Errorf("DeleteRawMotionByIDMock not set")
	}
	return fnsdc.DeleteRawMotionByIDMock(id)
}

func (fnsdc *FakeNoSQLDatabaseClient) GetRawMotion(userID string, timestamp time.Time) (*st.RawMotion, error) {
	if fnsdc.GetRawMotionMock == nil {
		return nil, fmt.Errorf("GetRawMotionMock not set")
	}
	return fnsdc.GetRawMotionMock(userID, timestamp)
}

func (fnsdc *FakeNoSQLDatabaseClient) InsertRawMotion(rawMotion *st.RawMotion) (string, error) {
	if fnsdc.InsertRawMotionMock == nil {
		return "", fmt.Errorf("InsertRawMotionMock not set")
	}
	return fnsdc.InsertRawMotionMock(rawMotion)
}

func (fnsdc *FakeNoSQLDatabaseClient) InsertUniqueRawMotion(rawMotion *st.RawMotion) (string, error) {
	if fnsdc.InsertUniqueRawMotionMock == nil {
		return "", fmt.Errorf("InsertUniqueRawMotionMock not set")
	}
	return fnsdc.InsertUniqueRawMotionMock(rawMotion)
}

func (fnsdc *FakeNoSQLDatabaseClient) ListAllUserIDs(pageToken string, maxResults int) ([]string, string, error) {
	if fnsdc.ListAllUserIDsMock == nil {
		return nil, "", fmt.Errorf("ListAllUserIDsMock not set")
	}
	return fnsdc.ListAllUserIDsMock(pageToken, maxResults)
}

func (fnsdc *FakeNoSQLDatabaseClient) GetUserByID(id string) (*st.User, error) {
	if fnsdc.GetUserByIDMock == nil {
		return nil, fmt.Errorf("ListAllUserIDsMock not set")
	}
	return fnsdc.GetUserByIDMock(id)
}

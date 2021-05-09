package rawmotiondao

import (
	"log"
	st "seneca/api/type"
)

type MockRawMotionDAO struct {
	InsertUniqueRawMotionMock func(rawMotion *st.RawMotion) (*st.RawMotion, error)
	GetRawMotionByIDMock      func(id string) (*st.RawMotion, error)
	ListUserRawMotionIDsMock  func(userID string) ([]string, error)
	DeleteRawMotionByIDMock   func(id string) error
}

func (mrld *MockRawMotionDAO) InsertUniqueRawMotion(rawMotion *st.RawMotion) (*st.RawMotion, error) {
	if mrld.InsertUniqueRawMotionMock == nil {
		log.Fatal("InsertUniqueRawMotionMock called but not set")
	}
	return mrld.InsertUniqueRawMotionMock(rawMotion)
}

func (mrld *MockRawMotionDAO) GetRawMotionByID(id string) (*st.RawMotion, error) {
	if mrld.GetRawMotionByIDMock == nil {
		log.Fatal("GetRawMotionByIDMock called but not set")
	}
	return mrld.GetRawMotionByIDMock(id)
}

func (mrld *MockRawMotionDAO) ListUserRawMotionIDs(userID string) ([]string, error) {
	if mrld.ListUserRawMotionIDsMock == nil {
		log.Fatal("ListUserRawMotionIDsMock called but not set")
	}
	return mrld.ListUserRawMotionIDsMock(userID)
}

func (mrld *MockRawMotionDAO) DeleteRawMotionByID(id string) error {
	if mrld.DeleteRawMotionByIDMock == nil {
		log.Fatal("DeleteRawMotionByIDMock called but not set")
	}
	return mrld.DeleteRawMotionByIDMock(id)
}

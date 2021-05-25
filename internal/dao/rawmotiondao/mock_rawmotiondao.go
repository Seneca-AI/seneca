package rawmotiondao

import (
	"context"
	"log"
	st "seneca/api/type"
)

type MockRawMotionDAO struct {
	InsertUniqueRawMotionMock       func(rawMotion *st.RawMotion) (*st.RawMotion, error)
	GetRawMotionByIDMock            func(id string) (*st.RawMotion, error)
	ListUserRawMotionIDsMock        func(userID string) ([]string, error)
	DeleteRawMotionByIDMock         func(id string) error
	PutRawMotionByIDMock            func(ctx context.Context, rawMotionID string, rawMotion *st.RawMotion) error
	ListUnprocessedRawMotionIDsMock func(userID string, latestVersion float64) ([]string, error)
}

func (mrmd *MockRawMotionDAO) InsertUniqueRawMotion(rawMotion *st.RawMotion) (*st.RawMotion, error) {
	if mrmd.InsertUniqueRawMotionMock == nil {
		log.Fatal("InsertUniqueRawMotionMock called but not set")
	}
	return mrmd.InsertUniqueRawMotionMock(rawMotion)
}

func (mrmd *MockRawMotionDAO) GetRawMotionByID(id string) (*st.RawMotion, error) {
	if mrmd.GetRawMotionByIDMock == nil {
		log.Fatal("GetRawMotionByIDMock called but not set")
	}
	return mrmd.GetRawMotionByIDMock(id)
}

func (mrmd *MockRawMotionDAO) ListUserRawMotionIDs(userID string) ([]string, error) {
	if mrmd.ListUserRawMotionIDsMock == nil {
		log.Fatal("ListUserRawMotionIDsMock called but not set")
	}
	return mrmd.ListUserRawMotionIDsMock(userID)
}

func (mrmd *MockRawMotionDAO) DeleteRawMotionByID(id string) error {
	if mrmd.DeleteRawMotionByIDMock == nil {
		log.Fatal("DeleteRawMotionByIDMock called but not set")
	}
	return mrmd.DeleteRawMotionByIDMock(id)
}

func (mrmd *MockRawMotionDAO) PutRawMotionByID(ctx context.Context, rawMotionID string, rawMotion *st.RawMotion) error {
	if mrmd.PutRawMotionByIDMock == nil {
		log.Fatal("PutRawMotionByIDMock called but not set")
	}
	return mrmd.PutRawMotionByIDMock(ctx, rawMotionID, rawMotion)
}

func (mrmd *MockRawMotionDAO) ListUnprocessedRawMotionIDs(userID string, latestVersion float64) ([]string, error) {
	if mrmd.ListUnprocessedRawMotionIDsMock == nil {
		log.Fatal("ListUnprocessedRawMotionIDsMock called but not set")
	}
	return mrmd.ListUnprocessedRawMotionIDsMock(userID, latestVersion)
}

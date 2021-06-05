package intraseneca

import (
	"fmt"
	"seneca/api/senecaerror"
	st "seneca/api/type"
)

type MockIntraSenecaClient struct {
	requestResponseMap map[string]map[string]interface{}
}

func NewMockIntraSenecaClient() *MockIntraSenecaClient {
	return &MockIntraSenecaClient{
		requestResponseMap: map[string]map[string]interface{}{},
	}
}

func (mc *MockIntraSenecaClient) ListTrips(req *st.TripListRequest) (*st.TripListResponse, error) {
	tripsListResponseMap, ok := mc.requestResponseMap[fmt.Sprintf("%T", req)]
	if !ok {
		return nil, fmt.Errorf("tripsListRequestMap is empty")
	}

	tripsListResponseObj, ok := tripsListResponseMap[makeTripsListRequestsKey(req)]
	if !ok {
		return nil, fmt.Errorf("no entry found in tripsListRequestMap for %s", makeTripsListRequestsKey(req))
	}

	tripsListResponse, ok := tripsListResponseObj.(*st.TripListResponse)
	if !ok {
		return nil, senecaerror.NewDevError(fmt.Errorf("want TripsListResponse, got %T", tripsListResponseObj))
	}

	return tripsListResponse, nil
}

func (mc *MockIntraSenecaClient) ProcessObjectsInVideo(req *st.ObjectsInFrameRequest) (*st.ObjectsInFrameResponse, error) {
	processObjectsResponseMap, ok := mc.requestResponseMap[fmt.Sprintf("%T", req)]
	if !ok {
		return nil, fmt.Errorf("processObjectsResponseMap is empty")
	}

	processObjectsResponseObj, ok := processObjectsResponseMap[req.RawFrame.Id]
	if !ok {
		return nil, fmt.Errorf("no entry found in processObjectsResponseMap for request with id: %s", req.RawFrame.Id)
	}

	processObjectsResponse, ok := processObjectsResponseObj.(*st.ObjectsInFrameResponse)
	if !ok {
		return nil, senecaerror.NewDevError(fmt.Errorf("want ObjectsInFrameResponse, got %T", processObjectsResponseObj))
	}

	return processObjectsResponse, nil
}

func (mc *MockIntraSenecaClient) InsertListTripsResponse(req *st.TripListRequest, resp *st.TripListResponse) {
	if _, ok := mc.requestResponseMap[fmt.Sprintf("%T", req)]; ok {
		mc.requestResponseMap[fmt.Sprintf("%T", req)] = map[string]interface{}{}
	}

	mc.requestResponseMap[fmt.Sprintf("%T", req)][makeTripsListRequestsKey(req)] = resp
}

func (mc *MockIntraSenecaClient) InsertProcessObjectsInFrameResponse(req *st.ObjectsInFrameRequest, resp *st.ObjectsInFrameResponse) {
	if _, ok := mc.requestResponseMap[fmt.Sprintf("%T", req)]; ok {
		mc.requestResponseMap[fmt.Sprintf("%T", req)] = map[string]interface{}{}
	}

	mc.requestResponseMap[fmt.Sprintf("%T", req)][req.RawFrame.Id] = resp
}

func makeTripsListRequestsKey(req *st.TripListRequest) string {
	return fmt.Sprintf("%s/%d/%d", req.UserId, req.StartTimeMs, req.EndTimeMs)
}

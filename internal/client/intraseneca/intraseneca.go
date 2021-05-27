package intraseneca

import st "seneca/api/type"

//nolint
type IntraSenecaInterface interface {
	ListTrips(req *st.TripListRequest) (*st.TripListResponse, error)
}

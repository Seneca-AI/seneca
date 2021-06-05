package intraseneca

import (
	st "seneca/api/type"
	"time"
)

type ServerConfig struct {
	SenecaServerHostName string
	SenecaServerHostPort string
	SenecaServerTimeout  time.Duration

	MLServerHostName string
	MLServerHostPort string
	MLServerTimeout  time.Duration
}

//nolint
type IntraSenecaInterface interface {
	ListTrips(req *st.TripListRequest) (*st.TripListResponse, error)
	ProcessObjectsInVideo(req *st.ObjectsInFrameRequest) (*st.ObjectsInFrameResponse, error)
}

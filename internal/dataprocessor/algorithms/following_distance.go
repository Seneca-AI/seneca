package algorithms

import (
	st "seneca/api/type"
	"seneca/internal/client/intraseneca"
)

type followingDistanceV0 struct {
	intraSenecaClient intraseneca.IntraSenecaInterface
}

func newFollowingDistanceV0(intraSenecaClient intraseneca.IntraSenecaInterface) *followingDistanceV0 {
	return &followingDistanceV0{
		intraSenecaClient: intraSenecaClient,
	}
}

func (fd *followingDistanceV0) GenerateEvents(inputs map[string][]interface{}) ([]*st.EventInternal, error) {
	return nil, nil
}

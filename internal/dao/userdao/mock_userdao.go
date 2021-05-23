package userdao

import (
	"log"
	st "seneca/api/type"
)

type MockUserDAO struct {
	InsertUniqueUserMock func(user *st.User) (*st.User, error)
	GetUserByIDMock      func(id string) (*st.User, error)
	ListAllUserIDsMock   func() ([]string, error)
	GetUserByEmailMock   func(email string) (*st.User, error)
}

func (mudao *MockUserDAO) InsertUniqueUser(user *st.User) (*st.User, error) {
	if mudao.InsertUniqueUserMock == nil {
		log.Fatalf("InsertUniqueUserMock called but not set.")
	}
	return mudao.InsertUniqueUserMock(user)
}

func (mudao *MockUserDAO) GetUserByID(id string) (*st.User, error) {
	if mudao.GetUserByIDMock == nil {
		log.Fatalf("GetUserByIDMock called but not set.")
	}
	return mudao.GetUserByIDMock(id)
}

func (mudao *MockUserDAO) ListAllUserIDs() ([]string, error) {
	if mudao.ListAllUserIDsMock == nil {
		log.Fatalf("ListAllUserIDsMock called but not set.")
	}
	return mudao.ListAllUserIDsMock()
}

func (mudao *MockUserDAO) GetUserByEmail(email string) (*st.User, error) {
	if mudao.GetUserByEmailMock == nil {
		log.Fatalf("GetUserByEmailMock called but not set.")
	}
	return mudao.GetUserByEmailMock(email)
}

package googledrive

import (
	"seneca/api/constants"
	st "seneca/api/type"
)

type UserClientFactory struct{}

func (fact *UserClientFactory) New(user *st.User) (GoogleDriveUserInterface, error) {
	return NewGoogleDriveUserClient(user, constants.PathToOAuthCredentials)
}

type clientOrError struct {
	client *FakeGoogleDriveUserClient
	err    error
}

type FakeUserClientFactory struct {
	clients map[string]*clientOrError
}

func NewFakeUserClientFactory() *FakeUserClientFactory {
	return &FakeUserClientFactory{
		clients: map[string]*clientOrError{},
	}
}

func (fakeFact *FakeUserClientFactory) New(user *st.User) (GoogleDriveUserInterface, error) {
	return fakeFact.clients[user.Id].client, fakeFact.clients[user.Id].err
}

func (fakeFact *FakeUserClientFactory) InsertFakeClient(userID string, fakeClient *FakeGoogleDriveUserClient, err error) {
	fakeFact.clients[userID] = &clientOrError{client: fakeClient, err: err}
}

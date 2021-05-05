package googledrive

import (
	"fmt"
	"os"
	st "seneca/api/type"
)

type UserClientFactory struct{}

func (fact *UserClientFactory) New(user *st.User) (GoogleDriveUserInterface, error) {
	pathToOAuthCredentials, exists := os.LookupEnv("GOOGLE_OAUTH_CREDENTIALS")
	if !exists {
		return nil, fmt.Errorf("GOOGLE_OAUTH_CREDENTIALS not set")
	}
	return NewGoogleDriveUserClient(user, pathToOAuthCredentials)
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

package userdao

import (
	"errors"
	"fmt"
	"seneca/api/constants"
	"seneca/api/senecaerror"
	st "seneca/api/type"
	"seneca/internal/client/database"
	"testing"
)

func TestInsertUniqueUser(t *testing.T) {
	user := &st.User{
		Email: "lucaloncar@gmail.com",
	}

	dao, sql := newUserDAOForTest()

	alreadyExistingUser := &st.User{
		Email: "lucaloncar@gmail.com",
	}

	// Already exists.
	existingID, err := sql.Create(constants.UsersTable, alreadyExistingUser)
	if err != nil {
		t.Fatalf("sql.Create(_, alreadyExistingUser) returns err: %v", err)
	}

	if _, err := dao.InsertUniqueUser(user); err == nil {
		t.Fatalf("Expected err for InsertUniqueUser() when already exists, got nil")
	}

	// No conflict with time email.
	alreadyExistingUser.Email = "some_email@gmail.com"
	if err := sql.Insert(constants.UsersTable, existingID, alreadyExistingUser); err != nil {
		t.Fatalf("DeleteByID() returns err: %v", err)
	}

	userWithID, err := dao.InsertUniqueUser(user)
	if err != nil {
		t.Fatalf("InsertUniqueUser() returns err: %v", err)
	}
	if userWithID.Id == "" {
		t.Fatalf("Newly created User not assigned ID")
	}

	someOtherUser := &st.User{
		Email: "something_else@gmail.com",
	}
	// Now induce errors for coverage.
	// 3 of them
	for i := 1; i < 4; i++ {
		sql.ErrorCalls = make(chan bool, 3)

		for j := 0; j < i; j++ {
			if j == i-1 {
				sql.ErrorCalls <- true
			} else {
				sql.ErrorCalls <- false
			}
		}

		if _, err := dao.InsertUniqueUser(someOtherUser); err == nil {
			t.Fatalf("Expected err from InsertUniqueUser() when call %d fails, got nil", i)
		}

		close(sql.ErrorCalls)
	}
}

func TestListUserUserIDs(t *testing.T) {
	dao, _ := newUserDAOForTest()

	wantUsers := []*st.User{}
	for i := 0; i < 20; i++ {
		user := &st.User{
			Id:    fmt.Sprintf("%d", i),
			Email: fmt.Sprintf("%d@mail.com", i),
		}

		userWithID, err := dao.InsertUniqueUser(user)
		if err != nil {
			t.Fatalf("InsertUniqueUser(%d) returns err: %v", i, err)
		}

		wantUsers = append(wantUsers, userWithID)
	}

	gotIDs, err := dao.ListAllUserIDs()
	if err != nil {
		t.Fatalf("ListAllUserIDs() returns err: %v", err)
	}

	if len(wantUsers) != len(gotIDs) {
		t.Fatalf("Want %d userIDs, got %d", len(wantUsers), len(gotIDs))
	}
}

func TestGetUserByEmail(t *testing.T) {
	user := &st.User{
		Email: "lucaloncar@gmail.com",
	}

	dao, sql := newUserDAOForTest()

	_, err := dao.GetUserByEmail("hello@mail.com")
	if err == nil {
		t.Fatalf("Expected err from GetUserByID() for non-existant ID, got nil")
	}
	var nfe *senecaerror.NotFoundError
	if !errors.As(err, &nfe) {
		t.Fatalf("Want NotFoundError from GetUserByID() for non-existant ID, got %v", err)
	}

	userWithID, err := dao.InsertUniqueUser(user)
	if err != nil {
		t.Fatalf("InsertUniqueUser() returns err: %v", err)
	}

	// No error.
	gotUser, err := dao.GetUserByEmail("lucaloncar@gmail.com")
	if err != nil {
		t.Fatalf("GetUserByEmail() returns err: %v", err)
	}
	if userWithID.Id != gotUser.Id {
		t.Fatalf("Users have same IDs but different values: %v - %v", userWithID, gotUser)
	}

	// Induce error.
	sql.ErrorCalls = make(chan bool, 1)
	sql.ErrorCalls <- true
	if _, err := dao.GetUserByID(gotUser.Id); err == nil {
		t.Fatalf("Expected err from GetUserByID() when call fails, got nil")
	}
	close(sql.ErrorCalls)
}

func TestGetUserByID(t *testing.T) {
	user := &st.User{
		Email: "lucaloncar@gmail.com",
	}

	dao, sql := newUserDAOForTest()

	_, err := dao.GetUserByID("456")
	if err == nil {
		t.Fatalf("Expected err from GetUserByID() for non-existant ID, got nil")
	}
	var nfe *senecaerror.NotFoundError
	if !errors.As(err, &nfe) {
		t.Fatalf("Want NotFoundError from GetUserByID() for non-existant ID, got %v", err)
	}

	userWithID, err := dao.InsertUniqueUser(user)
	if err != nil {
		t.Fatalf("InsertUniqueUser() returns err: %v", err)
	}

	// No error.
	gotUser, err := dao.GetUserByID(userWithID.Id)
	if err != nil {
		t.Fatalf("GetUserByID() returns err: %v", err)
	}
	if userWithID.Id != gotUser.Id {
		t.Fatalf("Users have same IDs but different values: %v - %v", userWithID, gotUser)
	}

	// Induce error.
	sql.ErrorCalls = make(chan bool, 1)
	sql.ErrorCalls <- true
	if _, err := dao.GetUserByID(userWithID.Id); err == nil {
		t.Fatalf("Expected err from GetUserByID() when call fails, got nil")
	}
	close(sql.ErrorCalls)
}

func newUserDAOForTest() (*SQLUserDAO, *database.FakeSQLDBService) {
	fakeSQLService := database.NewFake()
	return NewSQLUserDAO(fakeSQLService), fakeSQLService
}

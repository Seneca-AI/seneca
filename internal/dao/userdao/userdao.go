package userdao

import (
	"fmt"
	"seneca/api/constants"
	"seneca/api/senecaerror"
	st "seneca/api/type"
	"seneca/internal/client/cloud"
	"seneca/internal/dao"
)

const (
	emailFieldName = "Email"
)

type SQLUserDao struct {
	sqlInterface dao.SQLInterface
}

func NewSQLUserDao(sqlInterface dao.SQLInterface) *SQLUserDao {
	return &SQLUserDao{
		sqlInterface: sqlInterface,
	}
}

func (udao *SQLUserDao) InsertUniqueUser(user *st.User) (*st.User, error) {
	ids, err := udao.sqlInterface.ListIDs(constants.UsersTable, []*cloud.QueryParam{{FieldName: emailFieldName, Operand: "=", Value: user.Email}})
	if err != nil {
		return nil, fmt.Errorf("error listing users by email %q: %w", user.Email, err)
	}

	if len(ids) != 0 {
		return nil, fmt.Errorf("a user with email %q already exists", user.Email)
	}

	newUserID, err := udao.sqlInterface.Create(constants.UsersTable, user)
	if err != nil {
		return nil, fmt.Errorf("error inserting user %v into store: %w", user, err)
	}
	user.Id = newUserID

	// Now set the ID in the datastore object.
	if err := udao.sqlInterface.Insert(constants.UsersTable, user.Id, user); err != nil {
		return nil, fmt.Errorf("error updating userID for user %v - err: %w", user, err)
	}

	return user, nil
}

func (udao *SQLUserDao) GetUserByID(id string) (*st.User, error) {
	userObj, err := udao.sqlInterface.GetByID(constants.UsersTable, id)
	if err != nil {
		return nil, fmt.Errorf("sqlInterface.GetByID(%s, %s) returns err: %w", constants.UsersTable, id, err)
	}

	if userObj == nil {
		return nil, nil
	}

	user, ok := userObj.(*st.User)
	if !ok {
		return nil, fmt.Errorf("expected User, got %T", userObj)
	}
	return user, nil
}

func (udao *SQLUserDao) ListAllUserIDs() ([]string, error) {
	return udao.sqlInterface.ListIDs(constants.UsersTable, nil)
}

func (udao *SQLUserDao) GetUserByEmail(email string) (*st.User, error) {
	userIDs, err := udao.sqlInterface.ListIDs(constants.UsersTable, []*cloud.QueryParam{{FieldName: emailFieldName, Operand: "=", Value: email}})
	if err != nil {
		return nil, fmt.Errorf("error listing users with email %q - err: %w", email, err)
	}

	if len(userIDs) == 0 {
		return nil, senecaerror.NewNotFoundError(fmt.Errorf("no user with email %q found", email))
	}

	if len(userIDs) > 1 {
		return nil, senecaerror.NewBadStateError((fmt.Errorf("%d users with email %q found", len(userIDs), email)))
	}

	return udao.GetUserByID(userIDs[0])
}

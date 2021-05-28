package data

import (
	"fmt"
	"seneca/api/constants"
	"seneca/internal/client/database"
)

func DeleteAllUserDataInDB(userID string, includeUser bool, dbClient database.SQLInterface) error {
	for _, tableName := range constants.TableNames {
		if tableName == constants.UsersTable && includeUser {
			if err := dbClient.DeleteByID(tableName, userID); err != nil {
				return fmt.Errorf("DeleteByID(%s, %s) returns err: %w", tableName, userID, err)
			}
		}

		ids, err := dbClient.ListIDs(tableName, []*database.QueryParam{{FieldName: "UserId", Operand: "=", Value: userID}})
		if err != nil {
			return fmt.Errorf("ListIDs(%s, %s) returns err: %w", tableName, userID, err)
		}
		for _, id := range ids {
			if err := dbClient.DeleteByID(tableName, id); err != nil {
				return fmt.Errorf("DeleteByID(%s, %s) returns err: %w", tableName, id, err)
			}
		}
	}

	return nil
}

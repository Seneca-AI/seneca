package database

import (
	"seneca/api/constants"
	"time"
)

type SQLInterface interface {
	ListIDs(tableName constants.TableName, queryParams []*QueryParam) ([]string, error)
	GetByID(tableName constants.TableName, id string) (interface{}, error)
	Create(tableName constants.TableName, object interface{}) (string, error)
	Insert(tableName constants.TableName, id string, object interface{}) error
	DeleteByID(tableName constants.TableName, id string) error
}

type QueryParam struct {
	FieldName string
	Operand   string
	Value     interface{}
}

func GenerateTimeOffsetParams(fieldName string, createTimeMs int64, offset time.Duration) []*QueryParam {
	return []*QueryParam{
		{
			FieldName: fieldName,
			Operand:   ">=",
			Value:     createTimeMs - offset.Milliseconds(),
		},
		{
			FieldName: fieldName,
			Operand:   "<=",
			Value:     createTimeMs + offset.Milliseconds(),
		},
	}
}

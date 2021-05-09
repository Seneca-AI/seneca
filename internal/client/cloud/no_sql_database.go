package cloud

import (
	"time"
)

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

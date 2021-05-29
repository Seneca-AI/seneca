package database

import (
	"fmt"
	"log"
	"seneca/api/constants"
	st "seneca/api/type"
	"strings"
	"time"
)

type FakeSQLDBService struct {
	// key: (TableName, ID)
	data map[string]interface{}
	// The ErrorCalls chan is used to induce errors.  If false,
	// no error will be induced, if true, error will be induced.
	// See existing code for usage.
	ErrorCalls chan bool
}

func NewFake() *FakeSQLDBService {
	return &FakeSQLDBService{
		data:       map[string]interface{}{},
		ErrorCalls: nil,
	}
}

func (fs *FakeSQLDBService) ListIDs(tableName constants.TableName, queryParams []*QueryParam) ([]string, error) {
	if fs.ErrorCalls != nil {
		if <-fs.ErrorCalls {
			return nil, fmt.Errorf("errorMode")
		}
	}

	ids := []string{}
	for k, v := range fs.data {
		keyParts := strings.Split(k, "/")
		if len(keyParts) != 2 {
			log.Fatalf("Got %d key parts in FakeSQLDBService, must have 2", len(keyParts))
		}
		if keyParts[0] == tableName.String() && satisfiesQueryParams(tableName, v, queryParams) {
			ids = append(ids, keyParts[1])
		}
	}
	return ids, nil
}

func (fs *FakeSQLDBService) GetByID(tableName constants.TableName, id string) (interface{}, error) {
	if fs.ErrorCalls != nil {
		if <-fs.ErrorCalls {
			return nil, fmt.Errorf("errorMode")
		}
	}

	key := fmt.Sprintf("%s/%s", tableName.String(), id)
	obj, ok := fs.data[key]
	if !ok {
		return nil, nil
	}

	switch tableName {
	case constants.UsersTable:
		out, ok := obj.(*st.User)
		if !ok {
			log.Fatalf("got object of type %T for key %q", obj, key)
		}
		return out, nil
	case constants.RawVideosTable:
		out, ok := obj.(*st.RawVideo)
		if !ok {
			log.Fatalf("got object of type %T for key %q", obj, key)
		}
		return out, nil
	case constants.RawLocationsTable:
		out, ok := obj.(*st.RawLocation)
		if !ok {
			log.Fatalf("got object of type %T for key %q", obj, key)
		}
		return out, nil
	case constants.RawMotionsTable:
		out, ok := obj.(*st.RawMotion)
		if !ok {
			log.Fatalf("got object of type %T for key %q", obj, key)
		}
		return out, nil
	case constants.EventTable:
		out, ok := obj.(*st.EventInternal)
		if !ok {
			log.Fatalf("got object of type %T for key %q", obj, key)
		}
		return out, nil
	case constants.DrivingConditionTable:
		out, ok := obj.(*st.DrivingConditionInternal)
		if !ok {
			log.Fatalf("got object of type %T for key %q", obj, key)
		}
		return out, nil
	case constants.TripTable:
		out, ok := obj.(*st.TripInternal)
		if !ok {
			log.Fatalf("got object of type %T for key %q", obj, key)
		}
		return out, nil
	default:
		log.Fatalf("Invalid tableName %q", tableName)
	}
	return nil, nil
}

func (fs *FakeSQLDBService) Create(tableName constants.TableName, object interface{}) (string, error) {
	if fs.ErrorCalls != nil {
		if <-fs.ErrorCalls {
			return "", fmt.Errorf("errorMode")
		}
	}

	newID := fmt.Sprintf("%d", time.Now().UnixNano())
	key := fmt.Sprintf("%s/%s", tableName.String(), newID)
	fs.data[key] = object
	return newID, nil
}

func (fs *FakeSQLDBService) Insert(tableName constants.TableName, id string, object interface{}) error {
	if fs.ErrorCalls != nil {
		if <-fs.ErrorCalls {
			return fmt.Errorf("errorMode")
		}
	}

	key := fmt.Sprintf("%s/%s", tableName.String(), id)
	if _, ok := fs.data[key]; !ok {
		return fmt.Errorf("no value for key %q", key)
	}
	fs.data[key] = object
	return nil
}

func (fs *FakeSQLDBService) DeleteByID(tableName constants.TableName, id string) error {
	if fs.ErrorCalls != nil {
		if <-fs.ErrorCalls {
			return fmt.Errorf("errorMode")
		}
	}

	key := fmt.Sprintf("%s/%s", tableName.String(), id)
	if _, ok := fs.data[key]; !ok {
		return fmt.Errorf("no value for key %q", key)
	}
	delete(fs.data, key)
	return nil
}

func satisfiesQueryParams(tableName constants.TableName, object interface{}, queryParams []*QueryParam) bool {
	for _, qp := range queryParams {
		evaluateResult := func() bool {
			switch tableName {
			case constants.RawVideosTable:
				return evaluateOperand(getRawVideoField(qp.FieldName, object), qp.Value, qp.Operand)
			case constants.RawLocationsTable:
				return evaluateOperand(getRawLocationField(qp.FieldName, object), qp.Value, qp.Operand)
			case constants.RawMotionsTable:
				return evaluateOperand(getRawMotionField(qp.FieldName, object), qp.Value, qp.Operand)
			case constants.UsersTable:
				return evaluateOperand(getUserField(qp.FieldName, object), qp.Value, qp.Operand)
			case constants.TripTable:
				return evaluateOperand(getTripField(qp.FieldName, object), qp.Value, qp.Operand)
			case constants.EventTable:
				return evaluateOperand(getEventField(qp.FieldName, object), qp.Value, qp.Operand)
			case constants.DrivingConditionTable:
				return evaluateOperand(getDrivingConditionField(qp.FieldName, object), qp.Value, qp.Operand)
			default:
				log.Fatalf("satisfiesQueryParams() not yet implemented for table %q", tableName)
			}
			return false
		}()
		if !evaluateResult {
			return false
		}
	}
	return true
}

func getRawVideoField(fieldName constants.SenecaTypeFieldName, rawVideoObj interface{}) interface{} {
	rawVideo, ok := rawVideoObj.(*st.RawVideo)
	if !ok {
		log.Fatalf("Passed %T to getRawVideoField()", rawVideoObj)
	}

	switch fieldName {
	case constants.CreateTimeFieldName:
		return rawVideo.CreateTimeMs
	case constants.UserIDFieldName:
		return rawVideo.UserId
	case constants.AlgosVersionFieldName:
		return rawVideo.AlgosVersion
	default:
		log.Fatalf("Getting RawVideo field name %q not supported", fieldName)
	}
	return nil
}

func getRawLocationField(fieldName constants.SenecaTypeFieldName, rawLocationObj interface{}) interface{} {
	rawLocation, ok := rawLocationObj.(*st.RawLocation)
	if !ok {
		log.Fatalf("Passed %T to getRawLocationField()", rawLocationObj)
	}

	switch fieldName {
	case constants.TimestampFieldName:
		return rawLocation.TimestampMs
	case constants.UserIDFieldName:
		return rawLocation.UserId
	default:
		log.Fatalf("Getting RawLocation field name %q not supported", fieldName)
	}
	return nil
}

func getRawMotionField(fieldName constants.SenecaTypeFieldName, rawMotionObj interface{}) interface{} {
	rawMotion, ok := rawMotionObj.(*st.RawMotion)
	if !ok {
		log.Fatalf("Passed %T to getRawMotionField()", rawMotionObj)
	}

	switch fieldName {
	case constants.TimestampFieldName:
		return rawMotion.TimestampMs
	case constants.UserIDFieldName:
		return rawMotion.UserId
	case constants.AlgosVersionFieldName:
		return rawMotion.AlgosVersion
	default:
		log.Fatalf("Getting RawMotion field name %q not supported", fieldName)
	}
	return nil
}

func getUserField(fieldName constants.SenecaTypeFieldName, userObj interface{}) interface{} {
	user, ok := userObj.(*st.User)
	if !ok {
		log.Fatalf("Passed %T to getUserField()", userObj)
	}

	switch fieldName {
	case constants.EmailFieldName:
		return user.Email
	default:
		log.Fatalf("Getting User field name %q not supported", fieldName)
	}
	return nil
}

func getTripField(fieldName constants.SenecaTypeFieldName, tripObj interface{}) interface{} {
	trip, ok := tripObj.(*st.TripInternal)
	if !ok {
		log.Fatalf("Passed %T to getTripField()", tripObj)
	}

	switch fieldName {
	case constants.UserIDFieldName:
		return trip.UserId
	case constants.StartTimeFieldName:
		return trip.StartTimeMs
	case constants.EndTimeFieldName:
		return trip.EndTimeMs
	default:
		log.Fatalf("Getting TripInternal field name %q not supported", fieldName)
	}
	return nil
}

func getEventField(fieldName constants.SenecaTypeFieldName, eventObj interface{}) interface{} {
	event, ok := eventObj.(*st.EventInternal)
	if !ok {
		log.Fatalf("Passed %T to getEventField()", eventObj)
	}

	switch fieldName {
	case constants.UserIDFieldName:
		return event.UserId
	case constants.TripIDFieldName:
		return event.TripId
	default:
		log.Fatalf("Getting EventInternal field name %q not supported", fieldName)
	}
	return nil
}

func getDrivingConditionField(fieldName constants.SenecaTypeFieldName, drivingConditionObj interface{}) interface{} {
	drivingCondition, ok := drivingConditionObj.(*st.DrivingConditionInternal)
	if !ok {
		log.Fatalf("Passed %T to getDrivingConditionField()", drivingConditionObj)
	}

	switch fieldName {
	case constants.UserIDFieldName:
		return drivingCondition.UserId
	case constants.StartTimeFieldName:
		return drivingCondition.StartTimeMs
	case constants.EndTimeFieldName:
		return drivingCondition.EndTimeMs
	case constants.TripIDFieldName:
		return drivingCondition.TripId
	default:
		log.Fatalf("Getting DrivingConditionInternal field name %q not supported", fieldName)
	}
	return nil
}

var (
	stringType     = fmt.Sprintf("%T", "string")
	intType        = fmt.Sprintf("%T", int(0))
	int64Type      = fmt.Sprintf("%T", int64(0))
	float64Type    = fmt.Sprintf("%T", float64(0))
	boolType       = fmt.Sprintf("%T", false)
	supportedTypes = map[string]func(lhs interface{}, rhs interface{}, operand string) bool{
		stringType:  compareStrings,
		intType:     compareInts,
		int64Type:   compareInt64s,
		float64Type: compareFloat64s,
		// boolType will be caugh by "==" or "!=" below
		boolType: nil,
	}
)

func evaluateOperand(lhs interface{}, rhs interface{}, operand string) bool {
	if fmt.Sprintf("%T", lhs) != fmt.Sprintf("%T", rhs) {
		log.Fatalf("Attempting to evaluateOperand() on mismatched types %T and %T", lhs, rhs)
	}

	objType := fmt.Sprintf("%T", lhs)

	_, ok := supportedTypes[objType]
	if !ok {
		log.Fatalf("Attempting to evaluateOperand() on unsupported type %T", lhs)
	}

	switch operand {
	case "=":
		return lhs == rhs
	case "!=":
		return lhs != rhs
	default:
		function, ok := supportedTypes[objType]
		if !ok {
			log.Fatalf("Attempting to evaluateOperand() %q on type %T", operand, lhs)
		}
		return function(lhs, rhs, operand)
	}
}

func compareStrings(lhs interface{}, rhs interface{}, operand string) bool {
	lhsString, ok := lhs.(string)
	if !ok {
		log.Fatalf("Called compareStrings on type %T", lhs)
	}
	rhsString, ok := rhs.(string)
	if !ok {
		log.Fatalf("Called compareStrings on type %T", rhs)
	}

	switch operand {
	case ">":
		return lhsString > rhsString
	case "<":
		return lhsString < rhsString
	case ">=":
		return lhsString >= rhsString
	case "<=":
		return lhsString <= rhsString
	default:
		log.Fatalf("Unsupported operand %q on type %T", operand, lhs)
	}
	return false
}

func compareInt64s(lhs interface{}, rhs interface{}, operand string) bool {
	lhsInt64, ok := lhs.(int64)
	if !ok {
		log.Fatalf("Called compareInt64s on type %T", lhs)
	}
	rhsInt64, ok := rhs.(int64)
	if !ok {
		log.Fatalf("Called compareInt64s on type %T", rhs)
	}

	switch operand {
	case ">":
		return lhsInt64 > rhsInt64
	case "<":
		return lhsInt64 < rhsInt64
	case ">=":
		return lhsInt64 >= rhsInt64
	case "<=":
		return lhsInt64 <= rhsInt64
	default:
		log.Fatalf("Unsupported operand %q on type %T", operand, lhs)
	}
	return false
}

func compareInts(lhs interface{}, rhs interface{}, operand string) bool {
	lhsInt, ok := lhs.(int)
	if !ok {
		log.Fatalf("Called compareInts on type %T", lhs)
	}
	rhsInt, ok := rhs.(int)
	if !ok {
		log.Fatalf("Called compareInts on type %T", rhs)
	}
	return compareInt64s(int64(lhsInt), int64(rhsInt), operand)
}

func compareFloat64s(lhs interface{}, rhs interface{}, operand string) bool {
	lhsFloat64, ok := lhs.(float64)
	if !ok {
		log.Fatalf("Called compareFloat64s on type %T", lhs)
	}
	rhsFloat64, ok := rhs.(float64)
	if !ok {
		log.Fatalf("Called compareFloat64s on type %T", rhs)
	}

	switch operand {
	case ">":
		return lhsFloat64 > rhsFloat64
	case "<":
		return lhsFloat64 < rhsFloat64
	case ">=":
		return lhsFloat64 >= rhsFloat64
	case "<=":
		return lhsFloat64 <= rhsFloat64
	default:
		log.Fatalf("Unsupported operand %q on type %T", operand, lhs)
	}
	return false
}

package tests

import (
	"fmt"
	st "seneca/api/type"
	"seneca/internal/util"
	"seneca/test/integrationtest/testenv"
	"time"
)

func E2ERunner(testUserEmail string, testEnv *testenv.TestEnvironment) error {
	defer testEnv.Clean()

	user, err := testEnv.UserDAO.GetUserByEmail(testUserEmail)
	if err != nil {
		return fmt.Errorf("GetUserByEmail(%s) returns err: %w", testUserEmail, err)
	}

	accelerations := []int{5, 10, 15, 20}

	// Create a few raw motions.
	for i := 0; i < 12; i++ {
		rawMotion := &st.RawMotion{
			UserId: user.Id,
			Motion: &st.Motion{
				// 3 fast accelerations every period.
				AccelerationMphS: float64(accelerations[i%4]),
			},
			TimestampMs: util.TimeToMilliseconds(time.Date(2021, 05, 05, (i / 4), (i * 15 % 60), 0, 0, time.UTC)),
		}

		if _, err := testEnv.RawMotionDAO.InsertUniqueRawMotion(rawMotion); err != nil {
			return fmt.Errorf("InsertUniqueRawMotion() returns err: %v", err)
		}
	}

	// And a few raw videos.
	for i := 0; i < 3; i++ {
		rawVideo := &st.RawVideo{
			UserId:       user.Id,
			CreateTimeMs: util.TimeToMilliseconds(time.Date(2021, 05, 05, i, 0, 0, 0, time.UTC)),
			DurationMs:   int64(time.Minute * 1),
		}
		if _, err := testEnv.RawVideoDAO.InsertUniqueRawVideo(rawVideo); err != nil {
			return fmt.Errorf("InsertUniqueRawVideo() returns err: %v", err)
		}
	}

	testEnv.Runner.Run()

	tripIDs, err := testEnv.TripDAO.ListUserTripIDs(user.Id)
	if err != nil {
		return fmt.Errorf("ListUserTripIDs(%s) returns err: %w", testUserEmail, err)
	}
	if len(tripIDs) != 1 {
		return fmt.Errorf("want 1 tripID, got %d", len(tripIDs))
	}

	drivingConditionIDs, err := testEnv.DrivingConditionDAO.ListTripDrivingConditionIDs(user.Id, tripIDs[0])
	if err != nil {
		return fmt.Errorf("ListTripDrivingConditionIDs() returns err: %w", err)
	}
	if len(drivingConditionIDs) != 3 {
		return fmt.Errorf("want 3 drivingConditionIDs, got %d", len(drivingConditionIDs))
	}

	eventIDs, err := testEnv.EventDAO.ListTripEventIDs(user.Id, tripIDs[0])
	if err != nil {
		return fmt.Errorf("ListTripEventIDs() returns err: %w", err)
	}
	if len(eventIDs) != 9 {
		return fmt.Errorf("want 9 eventIDs, got %d", len(eventIDs))
	}

	return nil
}

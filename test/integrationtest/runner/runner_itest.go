package runner

import (
	"context"
	"fmt"
	st "seneca/api/type"
	"seneca/internal/controller/runner"
	"seneca/internal/dataprocessor"
	"seneca/internal/util"
	"seneca/test/integrationtest/testenv"
	"time"
)

func E2ERunner(testUserEmail string, testEnv *testenv.TestEnvironment) error {
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

	// Cleanup.
	defer func() {
		rawVideoIDs, err := testEnv.RawVideoDAO.ListUserRawVideoIDs(user.Id)
		if err != nil {
			testEnv.Logger.Error(fmt.Sprintf("ListUserRawVideoIDs(%s) returns err: %v", user.Id, err))
		}
		for _, rid := range rawVideoIDs {
			if err := testEnv.RawVideoDAO.DeleteRawVideoByID(rid); err != nil {
				testEnv.Logger.Error(fmt.Sprintf("DeleteRawVideoByID(%s) for user %q returns err: %v", rid, user.Id, err))
			}
		}

		rawMotionIDs, err := testEnv.RawMotionDAO.ListUserRawMotionIDs(user.Id)
		if err != nil {
			testEnv.Logger.Error(fmt.Sprintf("ListUserRawMotionIDs(%s) returns err: %v", user.Id, err))
		}
		for _, rlid := range rawMotionIDs {
			if err := testEnv.RawMotionDAO.DeleteRawMotionByID(rlid); err != nil {
				testEnv.Logger.Error(fmt.Sprintf("DeleteRawMotionByID(%s) for user %q returns err: %v", rlid, user.Id, err))
			}
		}

		tripIDs, err := testEnv.TripDAO.ListUserTripIDs(user.Id)
		if err != nil {
			testEnv.Logger.Error(fmt.Sprintf("ListUserTripIDs(%s) returns err: %v", user.Id, err))
		}
		for _, tid := range tripIDs {
			if err := testEnv.TripDAO.DeleteTripByID(context.TODO(), tid); err != nil {
				testEnv.Logger.Error(fmt.Sprintf("DeleteTripByID(%s) for user %q returns err: %v", tid, user.Id, err))
			}

			eventIDs, err := testEnv.EventDAO.ListTripEventIDs(user.Id, tid)
			if err != nil {
				testEnv.Logger.Error(fmt.Sprintf("ListTripEventIDs(%s) returns err: %v", user.Id, err))
			}
			for _, eid := range eventIDs {
				if err := testEnv.EventDAO.DeleteEventByID(context.TODO(), user.Id, tid, eid); err != nil {
					testEnv.Logger.Error(fmt.Sprintf("DeleteEventByID() for returns err: %v", err))
				}
			}

			drivingConditionIDs, err := testEnv.DrivingConditionDAO.ListTripDrivingConditionIDs(user.Id, tid)
			if err != nil {
				testEnv.Logger.Error(fmt.Sprintf("ListTripDrivingConditionIDs(%s) returns err: %v", user.Id, err))
			}
			for _, dcid := range drivingConditionIDs {
				if err := testEnv.DrivingConditionDAO.DeleteDrivingConditionByID(context.TODO(), user.Id, tid, dcid); err != nil {
					testEnv.Logger.Error(fmt.Sprintf("DeleteDrivingConditionByID() for returns err: %v", err))
				}
			}
		}
	}()

	dataprocessor := dataprocessor.New(dataprocessor.GetCurrentAlgorithms(), testEnv.EventDAO, testEnv.DrivingConditionDAO, testEnv.RawMotionDAO, testEnv.RawVideoDAO, testEnv.Logger)
	runner := runner.New(testEnv.UserDAO, dataprocessor, testEnv.Logger)
	runner.Run()

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

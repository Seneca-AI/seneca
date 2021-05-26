package tests

import (
	"fmt"
	"seneca/test/integrationtest/testenv"
	"time"
)

// E2ESource tests that the source is plumbed through all the way to the Trip.
func E2ESource(testUserEmail string, testEnv *testenv.TestEnvironment) error {
	defer testEnv.Clean()

	user, err := testEnv.UserDAO.GetUserByEmail(testUserEmail)
	if err != nil {
		return fmt.Errorf("GetUserByEmail(%s) returns err: %w", testUserEmail, err)
	}

	testEnv.Syncer.ScanAllUsers()
	testEnv.Runner.Run()

	trips, err := testEnv.APIServer.ListTrips(user.Id, time.Date(2021, 0, 0, 0, 0, 0, 0, time.UTC), time.Date(2022, 0, 0, 0, 0, 0, 0, time.UTC))
	if err != nil {
		return fmt.Errorf("ListTrips(%s) returns err: %w", user.Id, err)
	}

	if len(trips) == 0 {
		return fmt.Errorf("trips length is 0")
	}

	for _, trip := range trips {
		for _, event := range trip.Event {
			if event.ExternalSource == nil || event.ExternalSource.VideoUrl == "" {
				return fmt.Errorf("externalSource for event for user %q is %v", user.Id, event.ExternalSource)
			}
		}

		for _, dc := range trip.DrivingCondition {
			if len(dc.ExternalSource) == 0 {
				return fmt.Errorf("trip has 0 external sources for driving condition")
			}
			for _, es := range dc.ExternalSource {
				if es.VideoUrl == "" {
					return fmt.Errorf("external source has empty VideoURL")
				}
			}
		}
	}

	return nil
}

package env

import (
	"fmt"
	"os"
)

func ValidateEnvironmentVariables() error {
	_, exists := os.LookupEnv("GOOGLE_CLOUD_PROJECT")
	if !exists {
		return fmt.Errorf("GOOGLE_CLOUD_PROJECT not set")
	}

	_, exists = os.LookupEnv("GOOGLE_APPLICATION_CREDENTIALS")
	if !exists {
		return fmt.Errorf("GOOGLE_APPLICATION_CREDENTIALS not set")
	}

	_, exists = os.LookupEnv("GOOGLE_OAUTH_CREDENTIALS")
	if !exists {
		return fmt.Errorf("GOOGLE_OAUTH_CREDENTIALS not set")
	}

	return nil
}

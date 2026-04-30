package config

import (
	"fmt"
	"os"
)

func RequireEnv(keys ...string) error {
	for _, key := range keys {
		if os.Getenv(key) == "" {
			return fmt.Errorf("missing required env: %s", key)
		}
	}

	return nil
}

package seed

import (
	"errors"
	"fmt"
	"os"
)

var requiredEnvVars = []string{
	"POSTGRES_PASSWORD",
	"ADMIN_EMAIL",
	"ADMIN_PASSWORD",
	"AUTH_SECRET",
	"BASE_URL",
	"RESEND_API_KEY",
	"EMAIL_FROM",
	"OPENAI_API_KEY",
}

var ErrMissingEnv = errors.New("schoolrise: missing required environment variables")

func ValidateEnv() error {
	missing := []string{}

	for _, key := range requiredEnvVars {
		if os.Getenv(key) == "" {
			missing = append(missing, key)
		}
	}

	if len(missing) == 0 {
		return nil
	}

	return fmt.Errorf("%w: %v", ErrMissingEnv, missing)
}

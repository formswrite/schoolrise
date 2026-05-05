package secrets

import (
	"fmt"
	"os"
)

func MustEnv(name string) []byte {
	v := os.Getenv(name)
	if v == "" {
		panic(fmt.Sprintf("%s env var is required (no fallback for credentials); set it in .env or via your secret manager", name))
	}
	return []byte(v)
}

func MustEnvString(name string) string {
	return string(MustEnv(name))
}

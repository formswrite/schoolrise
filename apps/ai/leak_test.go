package ai_test

import (
	"testing"

	"go.uber.org/goleak"
)

func TestNoGoroutineLeaks(t *testing.T) {
	goleak.VerifyNone(t,
		goleak.IgnoreCurrent(),
	)
}

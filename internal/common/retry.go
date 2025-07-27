package common

import (
	"math"
	"math/rand"
	"time"
)

// RetryWithExponentialBackoff retries a function with exponential backoff
func RetryWithExponentialBackoff(fn func() error, maxRetries int, initialDelay time.Duration) error {
	var err error

	for i := 0; i < maxRetries; i++ {
		err = fn()
		if err == nil {
			return nil
		}

		// Calculate delay with exponential backoff
		delay := time.Duration(math.Pow(2, float64(i))) * initialDelay

		// Add jitter to prevent thundering herd
		jitter := time.Duration(rand.Int63n(int64(delay) / 2))
		delay += jitter

		time.Sleep(delay)
	}

	return err
}

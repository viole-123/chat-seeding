package retry

import (
	"fmt"
	"time"
)

// Do retries function with exponential backoff
func Do(fn func() error, maxRetries int, initialDelay time.Duration) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		err = fn()
		if err == nil {
			return nil // Success
		}

		if i < maxRetries-1 { // Don't sleep on last attempt
			delay := initialDelay * time.Duration(1<<uint(i)) // Exponential: 1s, 2s, 4s...
			time.Sleep(delay)
		}
	}
	return fmt.Errorf("failed after %d retries: %w", maxRetries, err)
}

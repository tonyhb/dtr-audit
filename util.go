package main

import "time"

// retry retries a function that produces an error up to 3 times with 5 second
// breaks before finally returning the last error trapped.
func retry(f func() error) (err error) {
	for i := 0; i < 3; i++ {
		// if there's no error break early and return nothing
		if err = f(); err == nil {
			return nil
		}

		// pause
		<-time.After(5 * time.Second)
	}

	// return the last error, or nil, caught by the function passed in
	return err
}

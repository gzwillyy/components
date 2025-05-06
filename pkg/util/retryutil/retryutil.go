package retryutil

import (
	"context"
	"errors"
	"time"
)

var (
	RetryableErr = errors.New("retry")
	TimeoutErr   = errors.New("timeout")
)

func RetryUntilTimeout(ctx context.Context, interval time.Duration, timeout time.Duration, do func() error) error {
	attempt := func() (done bool, err error) {
		err = do()
		if err == nil || err != RetryableErr {
			done = true
		}
		return
	}

	if done, err := attempt(); done {
		return err
	}

	if timeout == 0 {
		timeout = time.Duration(1<<63 - 1)
	}

	t := time.NewTimer(timeout)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
			return TimeoutErr
		case <-time.After(interval):
			if done, err := attempt(); done {
				return err
			}
		}
	}
}

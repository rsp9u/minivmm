package minivmm

import (
	"context"
	"os"
	"time"

	"github.com/gofrs/flock"
)

func WriteWithLock(f *os.File, lockpath string, data []byte) error {
	// NOTE: the lock file will not be removed.
	fileLock := flock.New(lockpath)
	lockCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	locked, err := fileLock.TryLockContext(lockCtx, 200*time.Millisecond)
	if err != nil {
		return err
	}
	if locked {
		f.Write(data)
		err = fileLock.Unlock()
		if err != nil {
			return err
		}
	}

	return nil
}

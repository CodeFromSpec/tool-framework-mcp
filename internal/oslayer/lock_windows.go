//go:build windows

// code-from-spec: SPEC/golang/implementation/oslayer/file/lock_windows@TEpbfFGJq7JwXvZYLF2FgR0Ucmk

package oslayer

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/sys/windows"
)

func fileLockShared(f *os.File, timeoutMs int) error {
	if timeoutMs <= 0 {
		ol := &windows.Overlapped{}
		err := windows.LockFileEx(
			windows.Handle(f.Fd()),
			windows.LOCKFILE_FAIL_IMMEDIATELY,
			0,
			^uint32(0),
			^uint32(0),
			ol,
		)
		if err != nil {
			return ErrLockTimeout
		}
		return nil
	}

	deadline := time.Now().Add(time.Duration(timeoutMs) * time.Millisecond)
	sleep := 1 * time.Millisecond

	for {
		ol := &windows.Overlapped{}
		err := windows.LockFileEx(
			windows.Handle(f.Fd()),
			windows.LOCKFILE_FAIL_IMMEDIATELY,
			0,
			^uint32(0),
			^uint32(0),
			ol,
		)
		if err == nil {
			return nil
		}
		if windows.ERROR_LOCK_VIOLATION != 0 && err == windows.ERROR_LOCK_VIOLATION {
			// lock held by another process, retry
		} else {
			return fmt.Errorf("%w: %w", ErrLockFailed, err)
		}
		if time.Now().After(deadline) || time.Now().Equal(deadline) {
			return ErrLockTimeout
		}
		time.Sleep(sleep)
		sleep *= 2
		if sleep > 100*time.Millisecond {
			sleep = 100 * time.Millisecond
		}
	}
}

func fileLockExclusive(f *os.File, timeoutMs int) error {
	if timeoutMs <= 0 {
		ol := &windows.Overlapped{}
		err := windows.LockFileEx(
			windows.Handle(f.Fd()),
			windows.LOCKFILE_FAIL_IMMEDIATELY|windows.LOCKFILE_EXCLUSIVE_LOCK,
			0,
			^uint32(0),
			^uint32(0),
			ol,
		)
		if err != nil {
			return ErrLockTimeout
		}
		return nil
	}

	deadline := time.Now().Add(time.Duration(timeoutMs) * time.Millisecond)
	sleep := 1 * time.Millisecond

	for {
		ol := &windows.Overlapped{}
		err := windows.LockFileEx(
			windows.Handle(f.Fd()),
			windows.LOCKFILE_FAIL_IMMEDIATELY|windows.LOCKFILE_EXCLUSIVE_LOCK,
			0,
			^uint32(0),
			^uint32(0),
			ol,
		)
		if err == nil {
			return nil
		}
		if windows.ERROR_LOCK_VIOLATION != 0 && err == windows.ERROR_LOCK_VIOLATION {
			// lock held by another process, retry
		} else {
			return fmt.Errorf("%w: %w", ErrLockFailed, err)
		}
		if time.Now().After(deadline) || time.Now().Equal(deadline) {
			return ErrLockTimeout
		}
		time.Sleep(sleep)
		sleep *= 2
		if sleep > 100*time.Millisecond {
			sleep = 100 * time.Millisecond
		}
	}
}

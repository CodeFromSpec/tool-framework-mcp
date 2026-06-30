//go:build windows

// code-from-spec: SPEC/golang/implementation/oslayer/file/lock_windows@SzCkZBKBNkkPIRl-1adTD1jtSYY

package oslayer

import (
	"errors"
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
		if errors.Is(err, windows.ERROR_LOCK_VIOLATION) {
		} else {
			return fmt.Errorf("%w: %w", ErrLockFailed, err)
		}
		if !time.Now().Before(deadline) {
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
		if errors.Is(err, windows.ERROR_LOCK_VIOLATION) {
		} else {
			return fmt.Errorf("%w: %w", ErrLockFailed, err)
		}
		if !time.Now().Before(deadline) {
			return ErrLockTimeout
		}
		time.Sleep(sleep)
		sleep *= 2
		if sleep > 100*time.Millisecond {
			sleep = 100 * time.Millisecond
		}
	}
}

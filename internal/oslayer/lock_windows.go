//go:build windows

// code-from-spec: SPEC/golang/implementation/oslayer/file/lock_windows@GFoa0p0dWT5fa-lMBPkOqe8WUXw

package oslayer

import (
	"errors"
	"fmt"
	"os"

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

	event, err := windows.CreateEvent(nil, 1, 0, nil)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrLockFailed, err)
	}
	defer windows.CloseHandle(event)

	ol := windows.Overlapped{
		HEvent: event,
	}

	err = windows.LockFileEx(
		windows.Handle(f.Fd()),
		0,
		0,
		^uint32(0),
		^uint32(0),
		&ol,
	)
	if err == nil {
		return nil
	}
	if errors.Is(err, windows.ERROR_IO_PENDING) {
		result, waitErr := windows.WaitForSingleObject(event, uint32(timeoutMs))
		if result == uint32(windows.WAIT_OBJECT_0) {
			return nil
		}
		if result == uint32(windows.WAIT_TIMEOUT) {
			windows.CancelIo(windows.Handle(f.Fd()))
			return ErrLockTimeout
		}
		return fmt.Errorf("%w: %w", ErrLockFailed, waitErr)
	}
	return fmt.Errorf("%w: %w", ErrLockFailed, err)
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

	event, err := windows.CreateEvent(nil, 1, 0, nil)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrLockFailed, err)
	}
	defer windows.CloseHandle(event)

	ol := windows.Overlapped{
		HEvent: event,
	}

	err = windows.LockFileEx(
		windows.Handle(f.Fd()),
		windows.LOCKFILE_EXCLUSIVE_LOCK,
		0,
		^uint32(0),
		^uint32(0),
		&ol,
	)
	if err == nil {
		return nil
	}
	if errors.Is(err, windows.ERROR_IO_PENDING) {
		result, waitErr := windows.WaitForSingleObject(event, uint32(timeoutMs))
		if result == uint32(windows.WAIT_OBJECT_0) {
			return nil
		}
		if result == uint32(windows.WAIT_TIMEOUT) {
			windows.CancelIo(windows.Handle(f.Fd()))
			return ErrLockTimeout
		}
		return fmt.Errorf("%w: %w", ErrLockFailed, waitErr)
	}
	return fmt.Errorf("%w: %w", ErrLockFailed, err)
}

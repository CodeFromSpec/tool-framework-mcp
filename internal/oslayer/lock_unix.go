//go:build !windows

// code-from-spec: SPEC/golang/implementation/oslayer/file/lock_unix@6gS0NevAggfu7KL_tGpiIZxeMCc

package oslayer

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"time"
)

func fileLockShared(f *os.File, timeoutMs int) error {
	return acquireLockLinux(f, syscall.LOCK_SH, timeoutMs)
}

func fileLockExclusive(f *os.File, timeoutMs int) error {
	return acquireLockLinux(f, syscall.LOCK_EX, timeoutMs)
}

func acquireLockLinux(f *os.File, flag int, timeoutMs int) error {
	if timeoutMs <= 0 {
		err := syscall.Flock(int(f.Fd()), flag|syscall.LOCK_NB)
		if err == nil {
			return nil
		}
		if errors.Is(err, syscall.EWOULDBLOCK) {
			return ErrLockTimeout
		}
		return fmt.Errorf("%w: %w", ErrLockFailed, err)
	}

	deadline := time.Now().Add(time.Duration(timeoutMs) * time.Millisecond)
	sleep := time.Millisecond

	for {
		err := syscall.Flock(int(f.Fd()), flag|syscall.LOCK_NB)
		if err == nil {
			return nil
		}
		if !errors.Is(err, syscall.EWOULDBLOCK) {
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

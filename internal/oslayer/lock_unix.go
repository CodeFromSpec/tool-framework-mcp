//go:build !windows

// code-from-spec: SPEC/golang/implementation/oslayer/file/lock_unix@61qPtDY-I8o7yuJ_kO0RuMLRq8M

package oslayer

import (
	"os"
	"syscall"
	"time"
)

func fileLockShared(f *os.File, timeoutMs int) error {
	return acquireLock(f, syscall.LOCK_SH, timeoutMs)
}

func fileLockExclusive(f *os.File, timeoutMs int) error {
	return acquireLock(f, syscall.LOCK_EX, timeoutMs)
}

func acquireLock(f *os.File, flag int, timeoutMs int) error {
	if timeoutMs <= 0 {
		err := syscall.Flock(int(f.Fd()), flag|syscall.LOCK_NB)
		if err == nil {
			return nil
		}
		if err == syscall.EWOULDBLOCK {
			return ErrLockTimeout
		}
		return err
	}

	deadline := time.Now().Add(time.Duration(timeoutMs) * time.Millisecond)
	sleep := time.Millisecond

	for {
		err := syscall.Flock(int(f.Fd()), flag|syscall.LOCK_NB)
		if err == nil {
			return nil
		}
		if err != syscall.EWOULDBLOCK {
			return err
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

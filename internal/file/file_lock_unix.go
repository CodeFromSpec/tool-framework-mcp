//go:build !windows

// code-from-spec: SPEC/golang/implementation/os/file/lock_unix@7uIVWlxtxgFeBvwLFatonXawjbE

package file

import (
	"os"
	"syscall"
)

func fileLockShared(f *os.File) error {
	return syscall.Flock(int(f.Fd()), syscall.LOCK_SH)
}

func fileLockExclusive(f *os.File) error {
	return syscall.Flock(int(f.Fd()), syscall.LOCK_EX)
}

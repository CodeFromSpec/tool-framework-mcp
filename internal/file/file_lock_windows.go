//go:build windows

// code-from-spec: SPEC/golang/implementation/os/file/lock_windows@sQb_W5w_pFGPf9-G71q7mlhl-0o

package file

import (
	"os"

	"golang.org/x/sys/windows"
)

func fileLockShared(f *os.File) error {
	ol := new(windows.Overlapped)
	err := windows.LockFileEx(windows.Handle(f.Fd()), 0, 0, ^uint32(0), ^uint32(0), ol)
	if err != nil {
		return err
	}
	return nil
}

func fileLockExclusive(f *os.File) error {
	ol := new(windows.Overlapped)
	err := windows.LockFileEx(windows.Handle(f.Fd()), windows.LOCKFILE_EXCLUSIVE_LOCK, 0, ^uint32(0), ^uint32(0), ol)
	if err != nil {
		return err
	}
	return nil
}

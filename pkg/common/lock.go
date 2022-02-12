package common

import (
	"os"
	"syscall"
)

// Lock should use with defer f.Unlock()
func Lock(lockPath string) (*os.File, error) {
	if ktLockFile, err := os.Open(lockPath); err != nil {
		return nil, err
	} else {
		return ktLockFile, syscall.Flock(int(ktLockFile.Fd()), syscall.LOCK_EX)
	}
}

// Unlock remove lock
func Unlock(ktLockFile *os.File) {
	_ = syscall.Flock(int(ktLockFile.Fd()), syscall.LOCK_UN)
}

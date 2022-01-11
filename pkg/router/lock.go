package router

import (
	"os"
	"syscall"
)

const pathKtLock = "/var/kt.lock"
var ktLockFile *os.File

// Lock should use with defer f.Unlock()
func Lock() (err error) {
	if ktLockFile, err = os.Open(pathKtLock); err != nil {
		return err
	} else {
		return syscall.Flock(int(ktLockFile.Fd()), syscall.LOCK_EX)
	}
}

// Unlock remove lock
func Unlock() {
	_ = syscall.Flock(int(ktLockFile.Fd()), syscall.LOCK_UN)
}

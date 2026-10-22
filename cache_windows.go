//go:build windows

package main

import "syscall"

func replaceCacheFile(from, to string) error {
	r, _, e := moveFile.Call(ptr(from), ptr(to), 1|8)
	if r == 0 {
		if e != syscall.Errno(0) {
			return e
		}
		return syscall.EINVAL
	}
	return nil
}

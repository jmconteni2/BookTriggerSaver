//go:build !windows

package main

import "os"

func replaceCacheFile(from, to string) error { return os.Rename(from, to) }

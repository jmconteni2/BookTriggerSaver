//go:build windows && !capture

package main

func captureLauncher() bool { return false }
func captureWantsFrame(*saverWindow) bool { return false }
func captureAfterPaint(*saverWindow) (bool, error) { return false, nil }

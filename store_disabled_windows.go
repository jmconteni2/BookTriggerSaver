//go:build windows && !store

package main

const storeMode = false

// These are never reached in a classic build; the release installer is unchanged.
func storeDataDir() (string, error) { return appDir() }
func storeMain(mode string, parent uintptr) error { return nil }

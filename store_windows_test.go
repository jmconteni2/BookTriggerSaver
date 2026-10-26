//go:build windows && store

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStoreLocalStateStableAcrossPackageVersion(t *testing.T) {
	first, err := storeLocalState(`C:\Users\Demo\AppData\Local`, "J.M.Contenidos.BookTriggerSaver_abc")
	if err != nil { t.Fatal(err) }
	second, err := storeLocalState(`C:\Users\Demo\AppData\Local`, "J.M.Contenidos.BookTriggerSaver_abc")
	if err != nil || first != second { t.Fatalf("unstable LocalState: %q / %q / %v", first, second, err) }
	if filepath.Base(first) != "LocalState" { t.Fatalf("wrong path: %q", first) }
}

func TestStoreSafePreviousMissingScreenSaver(t *testing.T) {
	old := backup{Version: 1, Values: map[string]regValue{
		"SCRNSAVE.EXE": stringValue(filepath.Join(t.TempDir(), "removed.scr")),
		"ScreenSaveActive": stringValue("1"),
		"ScreenSaveTimeOut": stringValue("600"),
	}}
	result := storeSafePrevious(old)
	if result.Values["SCRNSAVE.EXE"].Exists { t.Fatal("stale saver restored") }
	if stringOf(result.Values["ScreenSaveActive"]) != "0" { t.Fatal("missing saver remains active") }
	if stringOf(result.Values["ScreenSaveTimeOut"]) != "600" { t.Fatal("unrelated timeout changed") }
}

func TestStoreSafePreviousExistingScreenSaver(t *testing.T) {
	path := filepath.Join(t.TempDir(), "valid.scr")
	if err := os.WriteFile(path, []byte("file"), 0600); err != nil { t.Fatal(err) }
	old := backup{Version: 1, Values: map[string]regValue{
		"SCRNSAVE.EXE": stringValue(path),
		"ScreenSaveActive": stringValue("1"),
		"ScreenSaveTimeOut": stringValue("300"),
	}}
	result := storeSafePrevious(old)
	if stringOf(result.Values["SCRNSAVE.EXE"]) != path { t.Fatal("existing saver not retained") }
	if stringOf(result.Values["ScreenSaveActive"]) != "1" { t.Fatal("existing saver deactivated") }
}

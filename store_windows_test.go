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

func TestStorePackageQueryInstalled(t *testing.T) {
    tests := []struct { status uintptr; count uint32; installed bool; wantErr bool }{
        {0, 0, false, false},
        {0, 1, true, false},
        {122, 1, true, false},
        {122, 0, false, false},
        {15700, 0, false, false},
        {87, 0, false, true},
        {1168, 0, false, true},
    }
    for _, tt := range tests {
        got, err := storePackageQueryInstalled(tt.status, tt.count)
        if got != tt.installed || (err != nil) != tt.wantErr {
            t.Errorf("status=%d, count=%d => %v, %v; expected %v, err=%v", tt.status, tt.count, got, err, tt.installed, tt.wantErr)
        }
    }
}

func TestStoreFamilyValidation(t *testing.T) {
    for _, family := range []string{
        "J.M.Contenidos.BookTriggerSaver_abc",
        "J.M.Contenidos.BookTriggerSaver_x123",
    } {
        if err := storeValidateFamily(family); err != nil { t.Fatal(err) }
    }
    for _, family := range []string{
        "", "BookTriggerSaver_abc", "J.M.Contenidos.BookTriggerSaver_/evil", "J.M.Contenidos.BookTriggerSaver_..\\file",
    } {
        if err := storeValidateFamily(family); err == nil { t.Fatalf("accepted invalid family %q", family) }
    }
}

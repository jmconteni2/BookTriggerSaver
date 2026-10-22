//go:build windows

package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
	"unsafe"
)

// The probe is compiled only into the test binary; it never opens a URL.
func TestMain(m *testing.M) {
	if len(os.Args) == 3 && os.Args[1] == "--purchase-desktop-probe" {
		runtime.LockOSThread()
		thread, _, _ := kernel.NewProc("GetCurrentThreadId").Call()
		desktop, _, _ := user32.NewProc("GetThreadDesktop").Call(thread)
		var name [256]uint16
		var needed uint32
		ok, _, _ := user32.NewProc("GetUserObjectInformationW").Call(desktop, 2, uintptr(unsafe.Pointer(&name[0])), uintptr(unsafe.Sizeof(name)), uintptr(unsafe.Pointer(&needed)))
		if ok == 0 {
			os.Exit(2)
		}
		if os.WriteFile(os.Args[2], []byte(syscall.UTF16ToString(name[:])), 0600) != nil {
			os.Exit(3)
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}
func TestNativePurchaseHelperUsesDefaultDesktop(t *testing.T) {
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "desktop name.txt")
	process, err := startNormalDesktopProcess(exe, []string{"--purchase-desktop-probe", path})
	if err != nil {
		t.Fatal(err)
	}
	defer syscall.CloseHandle(process.Process)
	defer syscall.CloseHandle(process.Thread)
	result, err := syscall.WaitForSingleObject(process.Process, 15000)
	if err != nil || result != 0 {
		t.Fatalf("probe: result=%d err=%v", result, err)
	}
	name, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.EqualFold(string(name), "default") {
		t.Fatalf("helper inherited wrong desktop: %q", name)
	}
}
func TestPurchaseWaitAndNoDuplicateRetry(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if waitForNormalDesktop(ctx, func() bool { return false }) == nil {
		t.Fatal("locked desktop was ignored")
	}
	if err := waitForNormalDesktop(context.Background(), func() bool { return true }); err != nil {
		t.Fatal(err)
	}
	sentinel := errors.New("browser rejected request")
	if err := boundedBrowserOpen(context.Background(), func() error { return sentinel }); err != sentinel {
		t.Fatal("launch failure lost")
	}
	var calls atomic.Int32
	entered := make(chan struct{})
	release := make(chan struct{})
	done := make(chan error, 1)
	ctx, cancel = context.WithCancel(context.Background())
	go func() {
		done <- boundedBrowserOpen(ctx, func() error { calls.Add(1); close(entered); <-release; return nil })
	}()
	<-entered
	cancel()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("timeout not surfaced")
		}
	case <-time.After(time.Second):
		t.Fatal("launch blocked caller")
	}
	close(release)
	if calls.Load() != 1 {
		t.Fatal("browser launch retried")
	}
}
func TestPurchaseTraceDoesNotStoreFullURL(t *testing.T) {
	detail := purchaseTarget("https://example.test/buy?private=value")
	if strings.Contains(detail, "private") || strings.Contains(detail, "value") || !strings.Contains(detail, "host=example.test") {
		t.Fatalf("unsafe trace: %s", detail)
	}
}

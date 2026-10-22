//go:build windows

package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"
	"unsafe"
)

// Test-only child: records arguments/desktop and exits; no web is opened.
func TestMain(m *testing.M) {
	if len(os.Args) == 4 && os.Args[1] == "--explorer-launch-probe" {
		runtime.LockOSThread()
		thread, _, _ := kernel.NewProc("GetCurrentThreadId").Call()
		desktop, _, _ := user32.NewProc("GetThreadDesktop").Call(thread)
		var name [256]uint16
		var needed uint32
		ok, _, _ := user32.NewProc("GetUserObjectInformationW").Call(desktop, 2, uintptr(unsafe.Pointer(&name[0])), uintptr(unsafe.Sizeof(name)), uintptr(unsafe.Pointer(&needed)))
		if ok == 0 {
			os.Exit(2)
		}
		data, _ := json.Marshal([]string{syscall.UTF16ToString(name[:]), os.Args[3]})
		if os.WriteFile(os.Args[2], data, 0600) != nil {
			os.Exit(3)
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}
func TestNativeExplorerLaunchesOnDefaultDesktop(t *testing.T) {
	shellWindow, _, _ := user32.NewProc("GetShellWindow").Call()
	if shellWindow == 0 {
		t.Skip("Runner has no interactive Explorer; runtime browser validation still required")
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "explorer result.json")
	marker := "libro ñ & con espacios"
	args := "--explorer-launch-probe " + syscall.EscapeArg(path) + " " + syscall.EscapeArg(marker)
	err = onCOMThread(func() error {
		return withExplorerShell(func(shell uintptr) error { return explorerShellExecute(shell, exe, args) })
	})
	if err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		data, err := os.ReadFile(path)
		if err == nil {
			var result []string
			if json.Unmarshal(data, &result) != nil || len(result) != 2 || !strings.EqualFold(result[0], "default") || result[1] != marker {
				t.Fatalf("wrong desktop or changed arguments: %s", data)
			}
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("Explorer accepted launch but probe did not execute")
}
func TestPurchaseTraceDoesNotStoreFullURL(t *testing.T) {
	detail := purchaseTarget("https://example.test/buy?private=value")
	if strings.Contains(detail, "private") || strings.Contains(detail, "value") || !strings.Contains(detail, "host=example.test") {
		t.Fatal(detail)
	}
}

func TestNativeDispatchPreservesURLAndParameterOrder(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	type mockDispatch struct{ Table *[7]uintptr }
	var table [7]uintptr
	expected := "https://example.test/compra?asin=AB123&titulo=niño%20azul&origen=salva"
	calls := 0
	table[5] = syscall.NewCallback(func(this, iid, names, count, locale, out uintptr) uintptr {
		address := *(*unsafe.Pointer)(unsafe.Pointer(&out))
		*(*int32)(address) = 42
		return 0
	})
	table[6] = syscall.NewCallback(func(this, id, iid, locale, flags, params, out, exception, argError uintptr) uintptr {
		calls++
		address := *(*unsafe.Pointer)(unsafe.Pointer(&params))
		p := (*automationParams)(address)
		if id != 42 || flags != 1 || p.Count != 5 || p.Args == nil {
			t.Error("incorrect dispatch arguments")
			return 0x80004005
		}
		args := unsafe.Slice(p.Args, 5)
		if args[0].Type != 3 || args[0].Value != 1 {
			t.Error("wrong show value")
		}
		for n, want := range []string{"open", "", "", expected} {
			v := args[n+1]
			if v.Type != 8 {
				t.Error("expected BSTR")
				continue
			}
			length, _, _ := oleAutomation.NewProc("SysStringLen").Call(uintptr(v.Value))
			address := *(*unsafe.Pointer)(unsafe.Pointer(&v.Value))
			got := syscall.UTF16ToString(unsafe.Slice((*uint16)(address), int(length)))
			if got != want {
				t.Errorf("URL/argument changed: %q != %q", got, want)
			}
		}
		return 0
	})
	object := mockDispatch{Table: &table}
	if err := explorerShellExecute(uintptr(unsafe.Pointer(&object)), expected, ""); err != nil {
		t.Fatal(err)
	}
	runtime.KeepAlive(object)
	if calls != 1 {
		t.Fatalf("expected one request, got %d", calls)
	}
}

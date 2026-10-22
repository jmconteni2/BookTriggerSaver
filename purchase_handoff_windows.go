//go:build windows

package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

const saverVersion = "11.2"

func purchaseTrace(event, detail string) {
	root, err := appDir()
	if err != nil {
		return
	}
	path := filepath.Join(root, "apertura.log")
	if info, e := os.Stat(path); e == nil && info.Size() > 256<<10 {
		_ = os.Remove(path)
	}
	f, e := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if e == nil {
		defer f.Close()
		fmt.Fprintf(f, "%s version=%s pid=%d %s %s\n", time.Now().Format(time.RFC3339Nano), saverVersion, os.Getpid(), event, strings.ReplaceAll(detail, "\n", " "))
	}
}
func purchaseTarget(link string) string {
	u, _ := url.Parse(link)
	hash := sha256.Sum256([]byte(link))
	host := ""
	if u != nil {
		host = u.Hostname()
	}
	return fmt.Sprintf("host=%s id=%x", host, hash[:8])
}

// Use the same signed executable; no external launcher or script is installed.
// Bind its initial thread to the user's normal desktop, not ScreenSaver.
func handoffPurchase(link string) error {
	if !webURL(link) {
		return fmt.Errorf("El enlace de compra no es válido")
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	purchaseTrace("handoff_begin", purchaseTarget(link))
	process, err := startNormalDesktopProcess(exe, []string{"--open-purchase", strconv.Itoa(os.Getpid()), link})
	if err != nil {
		purchaseTrace("handoff_error", err.Error())
		return err
	}
	syscall.CloseHandle(process.Thread)
	syscall.CloseHandle(process.Process)
	purchaseTrace("handoff_created", fmt.Sprintf("helper=%d", process.ProcessId))
	return nil
}

func startNormalDesktopProcess(exe string, arguments []string) (syscall.ProcessInformation, error) {
	args := append([]string{exe}, arguments...)
	for i := range args {
		args[i] = syscall.EscapeArg(args[i])
	}
	command, err := syscall.UTF16PtrFromString(strings.Join(args, " "))
	var process syscall.ProcessInformation
	if err != nil {
		return process, err
	}
	startup := syscall.StartupInfo{Desktop: wide(`winsta0\default`), Flags: 0x80}
	startup.Cb = uint32(unsafe.Sizeof(startup))
	err = syscall.CreateProcess(wide(exe), command, nil, nil, false, 0x08000000, nil, nil, &startup, &process)
	return process, err
}

func waitForSaverExit(pid uint32) error {
	handle, err := syscall.OpenProcess(0x00100000, false, pid)
	if err == syscall.Errno(87) {
		return nil
	} // Parent already exited.
	if err != nil {
		return err
	}
	defer syscall.CloseHandle(handle)
	result, err := syscall.WaitForSingleObject(handle, 15000)
	if err != nil {
		return err
	}
	if result != 0 {
		return fmt.Errorf("El salvapantallas no ha terminado (espera %d)", result)
	}
	return nil
}
func normalDesktopActive() bool {
	h, _, _ := user32.NewProc("OpenInputDesktop").Call(0, 0, 1)
	if h == 0 {
		return false
	}
	defer user32.NewProc("CloseDesktop").Call(h)
	var name [256]uint16
	var needed uint32
	ok, _, _ := user32.NewProc("GetUserObjectInformationW").Call(h, 2, uintptr(unsafe.Pointer(&name[0])), uintptr(unsafe.Sizeof(name)), uintptr(unsafe.Pointer(&needed)))
	return ok != 0 && strings.EqualFold(syscall.UTF16ToString(name[:]), "default")
}

func waitForNormalDesktop(ctx context.Context, ready func() bool) error {
	for {
		if ready() {
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("No se ha desbloqueado el escritorio: %w", ctx.Err())
		case <-time.After(100 * time.Millisecond):
		}
	}
}
func boundedBrowserOpen(ctx context.Context, launch func() error) error {
	result := make(chan error, 1)
	go func() { result <- launch() }()
	select {
	case err := <-result:
		return err
	case <-ctx.Done():
		return fmt.Errorf("Windows no ha confirmado la apertura del navegador en 30 segundos. No se ha reenviado el enlace para evitar duplicados")
	}
}
func runPurchaseHelper(args []string) error {
	if len(args) != 3 || args[0] != "--open-purchase" || !webURL(args[2]) {
		return fmt.Errorf("Petición de compra no válida")
	}
	pid, err := strconv.ParseUint(args[1], 10, 32)
	if err != nil || pid == 0 || pid == uint64(os.Getpid()) {
		return fmt.Errorf("Proceso de origen no válido")
	}
	purchaseTrace("helper_started", purchaseTarget(args[2]))
	if err = waitForSaverExit(uint32(pid)); err != nil {
		purchaseTrace("parent_wait_error", err.Error())
		return err
	}
	purchaseTrace("saver_exited", "")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	if err = waitForNormalDesktop(ctx, normalDesktopActive); err != nil {
		purchaseTrace("desktop_wait_error", err.Error())
		return err
	}
	purchaseTrace("desktop_ready", "")
	openCtx, stop := context.WithTimeout(context.Background(), 30*time.Second)
	defer stop()
	started := time.Now()
	purchaseTrace("browser_request", purchaseTarget(args[2]))
	err = boundedBrowserOpen(openCtx, func() error { return openPurchase(args[2]) })
	if err != nil {
		purchaseTrace("browser_error", fmt.Sprintf("elapsed=%s %v", time.Since(started), err))
		return err
	}
	purchaseTrace("browser_accepted", fmt.Sprintf("elapsed=%s", time.Since(started)))
	return nil
}

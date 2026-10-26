//go:build windows && store

package main

import (
    "encoding/json"
    "errors"
    "fmt"
    "os"
    "os/exec"
    "strconv"
    "time"
    "unsafe"
    "path/filepath"
    "strings"
    "syscall"
)

const storeMode = true
const storeHelperPrefix = "--store-helper-"

// The Store application uses a stable, per-user LocalState directory.
func currentPackageFamily() (string, error) {
    p := kernel.NewProc("GetCurrentPackageFamilyName")
    var n uint32
    code, _, _ := p.Call(uintptr(unsafe.Pointer(&n)), 0)
    if code != 122 || n < 2 || n > 1024 { return "", fmt.Errorf("No se pudo identificar el paquete MSIX (%d)", code) }
    b := make([]uint16, n)
    code, _, _ = p.Call(uintptr(unsafe.Pointer(&n)), uintptr(unsafe.Pointer(&b[0])))
    if code != 0 { return "", fmt.Errorf("No se pudo leer la identidad MSIX (%d)", code) }
    return syscall.UTF16ToString(b), nil
}

func storeLocalState(local, family string) (string, error) {
    if local == "" || family == "" || strings.ContainsAny(family, "\\/:") { return "", errors.New("Carpeta de la aplicación no disponible") }
    return filepath.Join(local, "Packages", family, "LocalState"), nil
}

func storeDataDir() (string, error) {
    exe, err := os.Executable()
    if err != nil { return "", err }
    if strings.EqualFold(filepath.Ext(exe), ".scr") { return filepath.Dir(exe), nil }
    family, err := currentPackageFamily()
    if err != nil { return "", err }
    return storeLocalState(os.Getenv("LOCALAPPDATA"), family)
}

// This location is outside MSIX-managed LocalState. It must be created
// from the unpackaged helper, not the packaged launcher, to avoid AppData
// virtualization/removal on direct MSIX uninstall.
func storePersistentDir() (string, error) {
    local := os.Getenv("LOCALAPPDATA")
    if local == "" { return "", errors.New("LOCALAPPDATA no disponible") }
    return filepath.Join(local, "BookTriggerSaverStore"), nil
}

func storeSaverPath() (string, error) {
    d, err := storePersistentDir()
    if err != nil { return "", err }
    return filepath.Join(d, saverFile), nil
}

func storeValidateFamily(family string) error {
    if !strings.HasPrefix(family, "J.M.Contenidos.BookTriggerSaver_") ||
        strings.ContainsAny(family, "\\/:") || len(family) > 180 {
        return errors.New("La identidad del paquete no es válida")
    }
    return nil
}

// Run from the unpackaged LocalState bootstrapper. Keep the executable and
// backup present for a later fallback if the MSIX is directly uninstalled.
func storePreparePersistent(family string) error {
    if err := storeValidateFamily(family); err != nil { return err }
    source, err := os.Executable()
    if err != nil { return err }
    dir, err := storePersistentDir()
    if err != nil { return err }
    if err := os.MkdirAll(dir, 0700); err != nil { return err }
    dest := filepath.Join(dir, saverFile)
    data, err := os.ReadFile(source)
    if err != nil { return err }
    if err := atomicWrite(dest, data); err != nil { return err }
    return atomicWrite(filepath.Join(dir, "package-family.txt"), []byte(family))
}

// Query current-user installation from Windows, not a marker written by
// the packaged app. A missing MSIX can then be detected by the retained SCR.
func storePackageInstalled(family string) (bool, error) {
    if err := storeValidateFamily(family); err != nil { return false, err }
    var count, size uint32
    p := kernel.NewProc("GetPackagesByPackageFamily")
    status, _, _ := p.Call(ptr(family), uintptr(unsafe.Pointer(&count)), 0, uintptr(unsafe.Pointer(&size)), 0)
    switch status {
    case 0: // ERROR_SUCCESS; no matching packages when count is zero
        return count > 0, nil
    case 122: // ERROR_INSUFFICIENT_BUFFER: at least one installed package
        return count > 0, nil
    default:
        return false, fmt.Errorf("No se pudo comprobar la instalación MSIX (%d)", status)
    }
}

// If Windows launched the retained SCR after direct MSIX uninstall,
// restore the previous saver instead of displaying an uninstalled product.
func storeGuardPackage() (bool, error) {
    exe, err := os.Executable()
    if err != nil { return false, err }
    dir, err := storePersistentDir()
    if err != nil { return false, err }
    if !strings.EqualFold(filepath.Dir(exe), dir) { return false, nil }
    data, err := os.ReadFile(filepath.Join(dir, "package-family.txt"))
    if err != nil { return true, fmt.Errorf("No se puede verificar la instalación de BookTrigger: %w", err) }
    installed, err := storePackageInstalled(strings.TrimSpace(string(data)))
    if err != nil { return true, err }
    if installed { return false, nil }
    if err := storeDeactivateNative(); err != nil { return true, err }
    return true, nil
}

// Windows' idle screensaver host must run a stable Win32 .scr, not a file
// under the MSIX versioned WindowsApps directory.
func storeStageSaver() (string, error) {
    exe, err := os.Executable()
    if err != nil { return "", err }
    if strings.EqualFold(filepath.Ext(exe), ".scr") { return "", errors.New("La preparación debe abrirse desde BookTriggerSaver") }
    data, err := os.ReadFile(filepath.Join(filepath.Dir(exe), saverFile))
    if err != nil { return "", fmt.Errorf("No se encuentra BookTrigger.scr en el paquete: %w", err) }
    dir, err := storeDataDir()
    if err != nil { return "", err }
    if err := os.MkdirAll(dir, 0700); err != nil { return "", err }
    path := filepath.Join(dir, saverFile)
    if err := atomicWrite(path, data); err != nil { return "", err }
    return path, nil
}

func storeRunHelper(args ...string) error {
    path, err := storeSaverPath()
    if err != nil { return err }
    if err := exec.Command(path, args...).Run(); err != nil { return fmt.Errorf("Windows no completó la configuración: %w", err) }
    return nil
}

func storeActive() (bool, error) {
    path, err := storeStageSaver()
    if err != nil { return false, err }
    err = exec.Command(path, storeHelperPrefix+"check").Run()
    if err == nil { return true, nil }
    var exit *exec.ExitError
    if errors.As(err, &exit) && exit.ExitCode() == 5 { return false, nil }
    return false, fmt.Errorf("No se pudo comprobar el salvapantallas: %w", err)
}

func storeActiveNative() (bool, error) {
    path, err := storeSaverPath()
    if err != nil { return false, err }
    key, err := openKey(syscall.HKEY_CURRENT_USER, desktopKey, false)
    if err != nil { return false, err }
    defer syscall.RegCloseKey(key)
    value, err := readValue(key, "SCRNSAVE.EXE")
    if err != nil { return false, err }
    return strings.EqualFold(stringOf(value), path), nil
}

func storeActivate(seconds uint32) error {
    if seconds < 60 || seconds > 1800 { return errors.New("Tiempo de espera no válido") }
    if _, err := storeStageSaver(); err != nil { return err }
    family, err := currentPackageFamily()
    if err != nil { return err }
    return storeRunHelper(storeHelperPrefix+"activate", strconv.FormatUint(uint64(seconds), 10), family)
}

func storeActivateNative(seconds uint32, family string) error {
    if seconds < 60 || seconds > 1800 { return errors.New("Tiempo de espera no válido") }
    if err := checkPolicy(); err != nil { return err }
    if err := storePreparePersistent(family); err != nil { return err }
    dir, err := storePersistentDir()
    if err != nil { return err }
    path := filepath.Join(dir, saverFile)
    key, err := openKey(syscall.HKEY_CURRENT_USER, desktopKey, true)
    if err != nil { return err }
    defer syscall.RegCloseKey(key)
    previous, err := snapshot(key)
    if err != nil { return err }

    backupPath := filepath.Join(dir, "configuracion_anterior.json")
    backupExists := true
    if raw, err := os.ReadFile(backupPath); errors.Is(err, os.ErrNotExist) {
        backupExists = false
        saved, err := json.MarshalIndent(previous, "", "  ")
        if err != nil { return err }
        if err := atomicWrite(backupPath, saved); err != nil { return err }
    } else if err != nil {
        return err
    } else {
        var old backup
        if json.Unmarshal(raw, &old) != nil || old.Version != 1 || len(old.Values) != len(settings) {
            return errors.New("La copia anterior no es válida; no se sobrescribirá")
        }
    }

    if err := apply(key, path, seconds); err != nil {
        if !backupExists { _ = os.Remove(backupPath) }
        return errors.Join(err, restore(key, previous))
    }
    return nil
}

func storeDeactivate() error {
    if _, err := storeStageSaver(); err != nil { return err }
    return storeRunHelper(storeHelperPrefix+"deactivate")
}

// Do not reactivate a previously removed classic screensaver on uninstall.
func storeSafePrevious(old backup) backup {
    path := stringOf(old.Values["SCRNSAVE.EXE"])
    if path != "" {
        if info, err := os.Stat(path); err != nil || info.IsDir() {
            old.Values["SCRNSAVE.EXE"] = regValue{}
            old.Values["ScreenSaveActive"] = stringValue("0")
        }
    }
    return old
}

func storeDeactivateNative() error {
    dir, err := storePersistentDir()
    if err != nil { return err }
    path, err := storeSaverPath()
    if err != nil { return err }

    key, err := openKey(syscall.HKEY_CURRENT_USER, desktopKey, true)
    if err != nil { return err }
    defer syscall.RegCloseKey(key)
    current, err := readValue(key, "SCRNSAVE.EXE")
    if err != nil { return err }

    backupPath := filepath.Join(dir, "configuracion_anterior.json")
    if strings.EqualFold(stringOf(current), path) {
        raw, err := os.ReadFile(backupPath)
        if err != nil { return fmt.Errorf("No se encuentra la configuración anterior: %w", err) }
        var previous backup
        if json.Unmarshal(raw, &previous) != nil || previous.Version != 1 || len(previous.Values) != len(settings) {
            return errors.New("La copia anterior está dañada; no se cambiarán los ajustes")
        }
        if err := restore(key, storeSafePrevious(previous)); err != nil { return err }
    }
    if err := os.Remove(backupPath); err != nil && !errors.Is(err, os.ErrNotExist) { return err }
    message("BookTrigger está desactivado. Ya puedes desinstalarlo desde Microsoft Store o Configuración de Windows.", 0x40)
    return nil
}

// Helpers run from the staged, unpackaged .scr so the Windows shell and
// the configurator observe the same HKCU/Desktop state.
func storeNativeHelper(args []string) (bool, error) {
    if len(args) == 0 || !strings.HasPrefix(args[0], storeHelperPrefix) { return false, nil }
    exe, err := os.Executable()
    if err != nil { return true, err }
    if !strings.EqualFold(filepath.Ext(exe), ".scr") { return true, errors.New("El ayudante debe ser BookTrigger.scr") }
    switch strings.TrimPrefix(args[0], storeHelperPrefix) {
    case "check":
        active, err := storeActiveNative()
        if err != nil { return true, err }
        if !active { os.Exit(5) }
        return true, nil
    case "activate":
        if len(args) != 3 { return true, errors.New("Tiempo de activación ausente") }
        n, err := strconv.ParseUint(args[1], 10, 32)
        if err != nil { return true, err }
        return true, storeActivateNative(uint32(n), args[2])
    case "deactivate":
        return true, storeDeactivateNative()
    }
    return true, errors.New("Orden no reconocida")
}

func storeLogSaverStart() {
    dir, err := storeDataDir()
    if err != nil { return }
    f, err := os.OpenFile(filepath.Join(dir, "inicio_salvapantallas.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
    if err != nil { return }
    defer f.Close()
    fmt.Fprintf(f, "%s Inicio SCR pid=%d\n", time.Now().Format(time.RFC3339), os.Getpid())
}

func storeMain(mode string, parent uintptr) error {
    switch mode {
    case "saver", "preview":
        if storeMode {
            stopped, err := storeGuardPackage()
            if err != nil { return err }
            if stopped { return nil }
        }
        if mode == "saver" { storeLogSaverStart() }
        return runEngine(mode, parent)
    case "invalid-preview":
        return nil
    case "install", "config":
        active, err := storeActive()
        if err != nil { return err }
        if active {
            if dialog("BookTrigger", "Salvapantallas activo", "BookTrigger está seleccionado. Puedes cambiar el tiempo de espera o restaurar el salvapantallas que tenías antes. Para desinstalar la edición Microsoft Store, desactívala primero.", "Desactivar y restaurar") {
                return storeDeactivate()
            }
        }
        seconds, ok := installOptions()
        if !ok { return nil }
        if err := storeActivate(seconds); err != nil { return err }
        if completedPage(seconds) { return runEngine("saver", 0) }
        return nil
    case "uninstall":
        if dialog("BookTrigger", "Desactivar BookTrigger", "Se restaurará el salvapantallas anterior. Después puedes quitar la aplicación desde Configuración de Windows.", "Desactivar y restaurar") {
            return storeDeactivate()
        }
        return nil
    default:
        return errors.New("Opción no reconocida")
    }
}

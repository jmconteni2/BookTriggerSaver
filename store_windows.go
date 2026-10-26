//go:build windows && store

package main

import (
    "encoding/json"
    "errors"
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "syscall"
)

const storeMode = true

// The Store edition uses its own data and never registers a classic uninstaller.
func storeDataDir() (string, error) {
    root, err := os.UserConfigDir()
    if err != nil { return "", err }
    return filepath.Join(root, "BookTriggerSaverStore"), nil
}

// Windows invokes the packaged .scr directly for screensaver, preview and settings.
// Nothing is copied out of the MSIX package or installed to the classic app directory.
func storeSaverPath() (string, error) {
    exe, err := os.Executable()
    if err != nil { return "", err }
    path := filepath.Join(filepath.Dir(exe), saverFile)
    info, err := os.Stat(path)
    if err != nil { return "", fmt.Errorf("No se encuentra el salvapantallas en el paquete: %w", err) }
    if info.IsDir() { return "", errors.New("El salvapantallas del paquete no es un archivo") }
    return path, nil
}

func storeActive() (bool, error) {
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
    if err := checkPolicy(); err != nil { return err }
    path, err := storeSaverPath()
    if err != nil { return err }
    dir, err := storeDataDir()
    if err != nil { return err }
    if err := os.MkdirAll(dir, 0700); err != nil { return err }

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
    dir, err := storeDataDir()
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
        if err := restore(key, previous); err != nil { return err }
    }
    if err := os.Remove(backupPath); err != nil && !errors.Is(err, os.ErrNotExist) { return err }
    message("BookTrigger está desactivado. Ya puedes desinstalarlo desde Microsoft Store o Configuración de Windows.", 0x40)
    return nil
}

func storeMain(mode string, parent uintptr) error {
    switch mode {
    case "saver", "preview":
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

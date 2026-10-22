package main

import (
	"strconv"
	"strings"
)

// Parse only documented screen-saver modes. A missing/invalid preview handle
// never starts a full-screen saver inside the Windows settings dialog.
func parseMode(args []string, installed bool) (string, uintptr) {
	if len(args) == 0 {
		if installed {
			return "saver", 0
		}
		return "install", 0
	}
	a := strings.ToLower(strings.TrimSpace(args[0]))
	switch a {
	case "--install":
		return "install", 0
	case "--uninstall":
		return "uninstall", 0
	case "/s", "-s":
		return "saver", 0
	}
	if a == "/c" || a == "-c" || strings.HasPrefix(a, "/c:") || strings.HasPrefix(a, "-c:") {
		return "config", 0
	}
	if a == "/p" || a == "-p" || strings.HasPrefix(a, "/p:") || strings.HasPrefix(a, "-p:") {
		h := ""
		if i := strings.IndexByte(a, ':'); i >= 0 {
			h = strings.TrimSpace(a[i+1:])
		} else if len(args) > 1 {
			h = strings.TrimSpace(args[1])
		}
		n, e := strconv.ParseUint(h, 10, 64)
		if e != nil || n == 0 {
			return "invalid-preview", 0
		}
		return "preview", uintptr(n)
	}
	return "unknown", 0
}

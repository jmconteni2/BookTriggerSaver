//go:build windows

package main

import (
	"crypto/sha256"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const saverVersion = "11.5"

func purchaseTrace(event, detail string) {
	root, err := engineDataDir()
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

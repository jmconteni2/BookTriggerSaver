package main

import "testing"

func TestScreenSaverInvocation(t *testing.T) {
	for _, tc := range []struct {
		args      []string
		installed bool
		mode      string
		parent    uintptr
	}{
		{nil, false, "install", 0}, {nil, true, "saver", 0},
		{[]string{" /S "}, true, "saver", 0}, {[]string{"-s"}, false, "saver", 0},
		{[]string{"/P", "4294967297"}, true, "preview", 4294967297},
		{[]string{"/p:123"}, true, "preview", 123},
		{[]string{"/p"}, true, "invalid-preview", 0},
		{[]string{"/p:0"}, true, "invalid-preview", 0},
		{[]string{"/p:-2"}, true, "invalid-preview", 0},
		{[]string{"/p:garbage"}, true, "invalid-preview", 0},
		{[]string{"/c:123"}, true, "config", 0},
		{[]string{"--uninstall"}, true, "uninstall", 0},
		{[]string{"/something"}, false, "unknown", 0},
	} {
		mode, parent := parseMode(tc.args, tc.installed)
		if mode != tc.mode || parent != tc.parent {
			t.Errorf("%q: got %s/%d, want %s/%d", tc.args, mode, parent, tc.mode, tc.parent)
		}
	}
}

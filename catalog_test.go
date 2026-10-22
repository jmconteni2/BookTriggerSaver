package main

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

func pngSample() []byte {
	img := image.NewNRGBA(image.Rect(0, 0, 4, 6))
	img.Set(0, 0, color.NRGBA{120, 80, 30, 255})
	var b bytes.Buffer
	png.Encode(&b, img)
	return b.Bytes()
}
func TestCatalogKeepsPairsAndRejectsInvalidEntries(t *testing.T) {
	books := fixtureBooks()
	books = append(books, books[0], Book{Title: "Mala", Cover: "file:///tmp/book", Link: "https://example.test"}, Book{Title: "Sin enlace", Cover: "https://example.test/z", Link: "javascript:alert(1)"})
	raw, _ := json.Marshal(catalogReply{OK: true, Books: books})
	got, e := parseCatalog(raw)
	if e != nil || len(got) != 4 {
		t.Fatal(got, e)
	}
	if got[0].Title != books[0].Title || got[0].Cover != books[0].Cover || got[0].Link != books[0].Link {
		t.Fatal("metadata mismatched")
	}
	if got[3].Link != "" {
		t.Fatal("invalid URI accepted")
	}
	for _, raw := range [][]byte{[]byte(`{"ok":false,"libros":null}`), []byte(`<html>error</html>`), bytes.Repeat([]byte("x"), maxCatalogBytes+1)} {
		if _, e := parseCatalog(raw); e == nil {
			t.Fatal("invalid response accepted")
		}
	}
}
func TestNetworkDownloadCacheAndOfflineRecovery(t *testing.T) {
	var fail atomic.Bool
	pngData := pngSample()
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if fail.Load() {
			http.Error(w, "offline", 503)
			return
		}
		switch r.URL.Path {
		case "/catalog":
			json.NewEncoder(w).Encode(catalogReply{OK: true, Books: []Book{{Title: "Libro", Cover: server.URL + "/cover", Link: server.URL + "/buy"}}})
		case "/cover":
			w.Write(pngData)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	root := t.TempDir()
	s := newCatalogStore(root, server.URL+"/catalog")
	if e := s.refresh(context.Background()); e != nil {
		t.Fatal(e)
	}
	books, _, _ := s.snapshot()
	if len(books) != 1 || books[0].Link != server.URL+"/buy" || !validateImageFile(books[0].Path) {
		t.Fatal(books)
	}
	fail.Store(true)
	if e := s.refresh(context.Background()); e == nil {
		t.Fatal("expected network error")
	}
	got, _, _ := s.snapshot()
	if len(got) != 1 || got[0] != books[0] {
		t.Fatal("network failure erased working library")
	}
	recovered := newCatalogStore(root, server.URL+"/catalog")
	recovered.loadCache()
	got, _, _ = recovered.snapshot()
	if len(got) != 1 || got[0] != books[0] {
		t.Fatal("cache recovery failed")
	}
}
func TestMalformedImagesAndOversizedResponsesDoNotEnterCache(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/huge" {
			w.Write(bytes.Repeat([]byte{'x'}, 1025))
			return
		}
		w.Write([]byte("not an image"))
	}))
	defer server.Close()
	s := newCatalogStore(t.TempDir(), server.URL)
	if _, e := s.get(context.Background(), server.URL+"/huge", 1024); e == nil {
		t.Fatal("response limit ignored")
	}
	if _, ok := s.downloadImage(context.Background(), Book{Title: "Mal", Cover: server.URL + "/bad"}); ok {
		t.Fatal("malformed image accepted")
	}
	if _, e := os.Stat(filepath.Join(s.Root, "covers", imageName(server.URL+"/bad"))); !os.IsNotExist(e) {
		t.Fatal("bad file persisted")
	}
}
func TestCancelledDownloadsReturnAndCachePathsCannotEscape(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	s := newCatalogStore(t.TempDir(), "https://example.test")
	if _, e := s.get(ctx, s.Endpoint, 100); e == nil {
		t.Fatal("cancel ignored")
	}
	name := imageName("https://example.test/../../bad?name=x")
	if filepath.Base(name) != name || strings.Contains(name, "/") || len(name) != 68 {
		t.Fatal(name)
	}
}

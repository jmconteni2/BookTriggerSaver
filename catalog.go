package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const maxCatalogBytes = 4 << 20
const maxImageBytes = 12 << 20
const maxImagePixels = 16_000_000

type catalogReply struct {
	OK    bool   `json:"ok"`
	Books []Book `json:"libros"`
	Error string `json:"error"`
}

func webURL(s string) bool {
	u, e := url.Parse(s)
	return e == nil && (u.Scheme == "https" || u.Scheme == "http") && u.Hostname() != "" && u.User == nil
}
func cleanBooks(in []Book) []Book {
	seen := map[string]bool{}
	out := make([]Book, 0, len(in))
	for _, b := range in {
		b.Title = strings.TrimSpace(b.Title)
		b.Cover = strings.TrimSpace(b.Cover)
		b.Link = strings.TrimSpace(b.Link)
		if b.Title == "" || !webURL(b.Cover) {
			continue
		}
		if !webURL(b.Link) {
			b.Link = ""
		}
		id := bookID(b)
		if seen[id] {
			continue
		}
		seen[id] = true
		b.Path = ""
		out = append(out, b)
	}
	return out
}
func parseCatalog(data []byte) ([]Book, error) {
	if len(data) > maxCatalogBytes {
		return nil, errors.New("El catálogo supera el tamaño permitido")
	}
	var r catalogReply
	if e := json.Unmarshal(data, &r); e != nil {
		return nil, e
	}
	if !r.OK {
		return nil, errors.New("El servicio no ha podido consultar la biblioteca")
	}
	return cleanBooks(r.Books), nil
}
func imageName(url string) string {
	v := sha256.Sum256([]byte(url))
	return hex.EncodeToString(v[:]) + ".img"
}
func validateImageFile(path string) bool {
	f, e := os.Open(path)
	if e != nil {
		return false
	}
	defer f.Close()
	st, e := f.Stat()
	if e != nil || st.Size() < 16 || st.Size() > maxImageBytes {
		return false
	}
	c, _, e := image.DecodeConfig(f)
	return e == nil && c.Width > 0 && c.Height > 0 && int64(c.Width)*int64(c.Height) <= maxImagePixels
}
func saveAtomic(path string, data []byte) error {
	if e := os.MkdirAll(filepath.Dir(path), 0700); e != nil {
		return e
	}
	f, e := os.CreateTemp(filepath.Dir(path), "booktrigger-*")
	if e != nil {
		return e
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if _, e = f.Write(data); e == nil {
		e = f.Sync()
	}
	ce := f.Close()
	if e != nil {
		return e
	}
	if ce != nil {
		return ce
	}
	// Renaming within a directory is atomic on the supported local filesystems.
	return replaceCacheFile(tmp, path)
}

type CatalogStore struct {
	Root, Endpoint string
	Client         *http.Client
	mu             sync.RWMutex
	ready          []Book
	revision       uint64
	status         string
}

func newCatalogStore(root, endpoint string) *CatalogStore {
	return &CatalogStore{Root: root, Endpoint: strings.TrimSpace(endpoint), Client: &http.Client{Timeout: 20 * time.Second, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10 || !webURL(req.URL.String()) {
			return errors.New("Redirección no válida")
		}
		return nil
	}}, status: "Conectando con la biblioteca…"}
}
func (s *CatalogStore) snapshot() ([]Book, uint64, string) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]Book(nil), s.ready...), s.revision, s.status
}
func (s *CatalogStore) publish(books []Book, status string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ready = append([]Book(nil), books...)
	s.status = status
	s.revision++
}
func (s *CatalogStore) setStatus(status string) {
	s.mu.Lock()
	s.status = status
	s.revision++
	s.mu.Unlock()
}
func (s *CatalogStore) cached(books []Book) []Book {
	var out []Book
	for _, b := range books {
		b.Path = filepath.Join(s.Root, "covers", imageName(b.Cover))
		if validateImageFile(b.Path) {
			out = append(out, b)
		}
	}
	return out
}
func (s *CatalogStore) loadCache() {
	data, e := os.ReadFile(filepath.Join(s.Root, "catalog-v11.json"))
	if e != nil {
		return
	}
	books, e := parseCatalog(data)
	if e != nil {
		return
	}
	if ready := s.cached(books); len(ready) > 0 {
		s.publish(ready, "Biblioteca guardada · Conectando…")
	}
}
func (s *CatalogStore) get(ctx context.Context, u string, limit int64) ([]byte, error) {
	if !webURL(u) {
		return nil, errors.New("Dirección web no válida")
	}
	req, e := http.NewRequestWithContext(ctx, "GET", u, nil)
	if e != nil {
		return nil, e
	}
	req.Header.Set("User-Agent", "BookTriggerSaver/11.0")
	resp, e := s.Client.Do(req)
	if e != nil {
		return nil, e
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Respuesta HTTP %d", resp.StatusCode)
	}
	data, e := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if e != nil {
		return nil, e
	}
	if int64(len(data)) > limit {
		return nil, errors.New("La respuesta supera el tamaño permitido")
	}
	return data, nil
}
func (s *CatalogStore) downloadImage(ctx context.Context, b Book) (Book, bool) {
	b.Path = filepath.Join(s.Root, "covers", imageName(b.Cover))
	if validateImageFile(b.Path) {
		return b, true
	}
	data, e := s.get(ctx, b.Cover, maxImageBytes)
	if e != nil {
		return Book{}, false
	}
	// Validate before writing: malformed/oversized images never reach GDI+.
	cfg, _, e := image.DecodeConfig(bytes.NewReader(data))
	if e != nil || cfg.Width <= 0 || cfg.Height <= 0 || int64(cfg.Width)*int64(cfg.Height) > maxImagePixels {
		return Book{}, false
	}
	if _, _, e = image.Decode(bytes.NewReader(data)); e != nil {
		return Book{}, false
	}
	if e = saveAtomic(b.Path, data); e != nil {
		return Book{}, false
	}
	return b, true
}
func (s *CatalogStore) refresh(ctx context.Context) error {
	data, e := s.get(ctx, s.Endpoint, maxCatalogBytes)
	if e != nil {
		return e
	}
	books, e := parseCatalog(data)
	if e != nil {
		return e
	}
	// An empty or temporarily broken feed must not erase the working library.
	if len(books) == 0 {
		return errors.New("El catálogo está temporalmente vacío")
	}
	if e = os.MkdirAll(filepath.Join(s.Root, "covers"), 0700); e != nil {
		return e
	}
	jobs := make(chan Book)
	results := make(chan Book)
	var workers sync.WaitGroup
	for i := 0; i < 4; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for b := range jobs {
				if b, ok := s.downloadImage(ctx, b); ok {
					select {
					case results <- b:
					case <-ctx.Done():
						return
					}
				}
			}
		}()
	}
	go func() {
		defer close(jobs)
		for _, b := range books {
			select {
			case jobs <- b:
			case <-ctx.Done():
				return
			}
		}
	}()
	go func() { workers.Wait(); close(results) }()
	ready := make([]Book, 0, len(books))
	previous, _, _ := s.snapshot()
	lastPublish := time.Now()
	for b := range results {
		ready = append(ready, b)
		if len(previous) == 0 && (len(ready) >= min(8, len(books)) && time.Since(lastPublish) > 500*time.Millisecond) {
			s.publish(ready, "Cargando la biblioteca…")
			lastPublish = time.Now()
		}
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if len(ready) == 0 {
		return errors.New("No se han podido cargar las portadas")
	}
	// Persist only successfully decoded books, keeping each original purchase URL.
	reply := catalogReply{OK: true, Books: ready}
	saved, _ := json.Marshal(reply)
	if e = saveAtomic(filepath.Join(s.Root, "catalog-v11.json"), saved); e != nil {
		s.publish(ready, "Biblioteca online · Caché no disponible")
		return nil
	}
	s.publish(ready, "Biblioteca online")
	return nil
}
func (s *CatalogStore) run(ctx context.Context) {
	s.loadCache()
	for {
		if e := s.refresh(ctx); e != nil && ctx.Err() == nil {
			ready, _, _ := s.snapshot()
			if len(ready) > 0 {
				s.setStatus("Biblioteca guardada · Sin conexión")
			} else {
				s.setStatus("Sin conexión · Reintentando…")
			}
		}
		t := time.NewTimer(2 * time.Minute)
		select {
		case <-ctx.Done():
			t.Stop()
			return
		case <-t.C:
		}
	}
}

package updatersource

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// testSource builds a Source against httptest TLS servers with the
// production bounds shrunk. transportServer is the server whose client
// transport (which trusts ITS cert) the Source dials with — for the two-
// server redirect test that is the asset server. Hosts are the servers' own
// 127.0.0.1 so the host allow-list logic itself is exercised, never bypassed.
func testSource(t *testing.T, api, asset, transportServer *httptest.Server, maxBytes int64, maxRedirects int) *Source {
	t.Helper()
	cfg := Config{
		Hosts:        []string{"127.0.0.1"},
		MaxBytes:     maxBytes,
		Timeout:      5 * time.Second,
		CacheTTL:     time.Minute,
		MaxRedirects: maxRedirects,
		APIBase:      api.URL,
		DownloadBase: asset.URL,
		Transport:    transportServer.Client().Transport,
	}
	return New(cfg)
}

// TestAllowListPinned pins the real hosts (fold A3): the asset redirect host
// was verified live 2026-10-02 [A] — github.com → 302 →
// release-assets.githubusercontent.com.
func TestAllowListPinned(t *testing.T) {
	want := map[string]bool{
		"api.github.com":                       true,
		"github.com":                           true,
		"objects.githubusercontent.com":        true,
		"release-assets.githubusercontent.com": true,
	}
	for _, h := range DefaultHosts {
		if !want[h] {
			t.Fatalf("DefaultHosts contains %q, not in the pinned set", h)
		}
	}
	for h := range want {
		if !AllowedHost(h) {
			t.Fatalf("pinned host %q missing from the allow-list", h)
		}
	}
	if AllowedHost("evil.example.com") {
		t.Fatal("off-list host reported allowed")
	}
}

func TestCheckURLRefusesHTTPAndOffList(t *testing.T) {
	if err := checkURL(&url.URL{Scheme: "http", Host: "api.github.com"}, DefaultHosts); err == nil {
		t.Fatal("http scheme accepted")
	}
	if err := checkURL(&url.URL{Scheme: "https", Host: "evil.example.com"}, DefaultHosts); err == nil {
		t.Fatal("off-list host accepted")
	}
	if err := checkURL(&url.URL{Scheme: "https", Host: "api.github.com"}, DefaultHosts); err != nil {
		t.Fatalf("allow-listed host refused: %v", err)
	}
}

func TestLatestRateLimitedIsCached(t *testing.T) {
	var hits int32
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()
	s := testSource(t, srv, srv, srv, 1<<20, 3)
	_, err := s.Latest(context.Background())
	if !errors.Is(err, ErrRateLimited) {
		t.Fatalf("first answer: %v, want ErrRateLimited", err)
	}
	_, err = s.Latest(context.Background())
	if !errors.Is(err, ErrRateLimited) {
		t.Fatalf("cached answer: %v, want ErrRateLimited", err)
	}
	if got := atomic.LoadInt32(&hits); got != 1 {
		t.Fatalf("API hits = %d, want 1 (the second answer is the cache)", got)
	}
}

func TestLatestParsesTagAndCommitish(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"tag_name":"v2026.10.01.1","target_commitish":"`+strings.Repeat("ab", 20)+`"}`)
	}))
	defer srv.Close()
	s := testSource(t, srv, srv, srv, 1<<20, 3)
	latest, err := s.Latest(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if latest.Tag != "v2026.10.01.1" || latest.TargetCommitish != strings.Repeat("ab", 20) {
		t.Fatalf("latest = %+v", latest)
	}
}

func TestDownloadRefusesContentLengthOverCap(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "2048")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	dir := t.TempDir()
	s := testSource(t, srv, srv, srv, 1024, 3)
	if _, err := s.Download(context.Background(), "v9.9.9", dir); !errors.Is(err, ErrSizeCap) {
		t.Fatalf("err = %v, want ErrSizeCap", err)
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 0 {
		t.Fatalf("destination not empty after a refused download: %v", entries)
	}
}

func TestDownloadRefusesStreamingOverCapAndRemovesPartial(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// No Content-Length: the stream itself must be capped.
		w.WriteHeader(http.StatusOK)
		chunk := make([]byte, 512)
		for i := 0; i < 8; i++ { // 4 KiB > cap 1024
			if _, err := w.Write(chunk); err != nil {
				return
			}
		}
	}))
	defer srv.Close()
	dir := t.TempDir()
	s := testSource(t, srv, srv, srv, 1024, 3)
	if _, err := s.Download(context.Background(), "v9.9.9", dir); !errors.Is(err, ErrSizeCap) {
		t.Fatalf("err = %v, want ErrSizeCap", err)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 0 {
		t.Fatalf("partial file left behind: %v", entries)
	}
}

func TestDownloadRefusesOffListRedirect(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// http + off-list host: refused by the redirect policy before a dial.
		http.Redirect(w, r, "http://localhost:1/x", http.StatusFound)
	}))
	defer srv.Close()
	dir := t.TempDir()
	s := testSource(t, srv, srv, srv, 1024, 3)
	if _, err := s.Download(context.Background(), "v9.9.9", dir); err == nil {
		t.Fatal("off-list redirect accepted")
	}
}

func TestDownloadRefusesTooManyRedirects(t *testing.T) {
	var srv *httptest.Server
	srv = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, srv.URL+"/hop", http.StatusFound)
	}))
	defer srv.Close()
	dir := t.TempDir()
	s := testSource(t, srv, srv, srv, 1024, 3)
	if _, err := s.Download(context.Background(), "v9.9.9", dir); err == nil {
		t.Fatal("redirect loop past MaxRedirects accepted")
	}
}

func TestDownloadFollowsAllowedRedirectsAndWritesFile(t *testing.T) {
	asset := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "release tarball bytes")
	}))
	defer asset.Close()
	api := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, asset.URL+"/asset", http.StatusFound)
	}))
	defer api.Close()
	// TWO servers = two self-signed certs; this test-only transport trusts
	// both so the hop from api to asset can complete. The production client
	// never sees InsecureSkipVerify (it is a test-local Config injection).
	testTransport := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	dir := t.TempDir()
	cfg := Config{
		Hosts:        []string{"127.0.0.1"},
		MaxBytes:     1 << 20,
		Timeout:      5 * time.Second,
		CacheTTL:     time.Minute,
		MaxRedirects: 3,
		APIBase:      api.URL,
		DownloadBase: api.URL,
		Transport:    testTransport,
	}
	s := New(cfg)
	got, err := s.Download(context.Background(), "v9.9.9", dir)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(dir, "v9.9.9.tar.gz")
	if got != want {
		t.Fatalf("path = %q, want %q", got, want)
	}
	b, err := os.ReadFile(got)
	if err != nil || string(b) != "release tarball bytes" {
		t.Fatalf("body = %q, %v", b, err)
	}
}

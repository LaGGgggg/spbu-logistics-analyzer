package osmdata

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var downloads sync.Map

const maximumDownloadBytes int64 = 4 << 30

func ensureExtract(ctx context.Context, files Files) error {
	lock, _ := downloads.LoadOrStore(files.OSM, make(chan struct{}, 1))
	semaphore := lock.(chan struct{})
	select {
	case semaphore <- struct{}{}:
	case <-ctx.Done():
		return ctx.Err()
	}
	defer func() { <-semaphore }()
	if _, err := os.Stat(files.OSM); err == nil {
		return verifyChecksum(ctx, files.OSM, files.SHA256)
	} else if !os.IsNotExist(err) {
		return err
	}
	if files.URL == "" {
		return fmt.Errorf("OSM file %s is missing; configure an extract URL or install the regional file", files.OSM)
	}
	u, err := url.Parse(files.URL)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
		return fmt.Errorf("invalid extract URL")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, files.URL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "SPBU-Logistics-Analyzer/1.0")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("extract download: HTTP %d", resp.StatusCode)
	}
	if resp.ContentLength > maximumDownloadBytes {
		return fmt.Errorf("extract exceeds download limit")
	}
	if err := os.MkdirAll(filepath.Dir(files.OSM), 0755); err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(files.OSM), ".extract-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	n, copyErr := io.Copy(f, io.LimitReader(resp.Body, maximumDownloadBytes+1))
	closeErr := f.Close()
	if copyErr != nil {
		return copyErr
	}
	if closeErr != nil {
		return closeErr
	}
	if n > maximumDownloadBytes {
		return fmt.Errorf("extract exceeds download limit")
	}
	if err := verifyChecksum(ctx, f.Name(), files.SHA256); err != nil {
		return err
	}
	return os.Rename(f.Name(), files.OSM)
}
func verifyChecksum(ctx context.Context, path, want string) error {
	if want == "" {
		return ctx.Err()
	}
	if decoded, err := hex.DecodeString(want); err != nil || len(decoded) != sha256.Size {
		return fmt.Errorf("invalid SHA256 in source catalog")
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	hash := sha256.New()
	buffer := make([]byte, 128*1024)
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		n, err := f.Read(buffer)
		if n > 0 {
			_, _ = hash.Write(buffer[:n])
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	if !strings.EqualFold(hex.EncodeToString(hash.Sum(nil)), want) {
		return fmt.Errorf("extract SHA256 mismatch")
	}
	return nil
}

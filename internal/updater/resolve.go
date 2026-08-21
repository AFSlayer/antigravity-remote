package updater

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"runtime"
	"strings"
	"time"
)

const DefaultDownloadPage = "https://antigravity.google/download"

var (
	hubURLRe     = regexp.MustCompile(`https://storage\.googleapis\.com/antigravity-public/antigravity-hub/[^"'<> ]+`)
	hubVersionRe = regexp.MustCompile(`/antigravity-hub/([0-9][0-9.]*)-[0-9]+/`)
)

// PlatformSlug maps runtime.GOOS and runtime.GOARCH to the official Antigravity hub artifact path.
func PlatformSlug(goos, goarch string) (string, error) {
	switch goos {
	case "linux":
		switch goarch {
		case "amd64":
			return "linux-x64/Antigravity.tar.gz", nil
		case "arm64":
			return "linux-arm/Antigravity.tar.gz", nil
		}
	case "darwin":
		switch goarch {
		case "arm64":
			return "darwin-arm/Antigravity.dmg", nil
		case "amd64":
			return "darwin-x64/Antigravity.dmg", nil
		}
	case "windows":
		switch goarch {
		case "amd64":
			return "windows-x64/Antigravity-x64.exe", nil
		case "arm64":
			return "windows-arm/Antigravity-arm64.exe", nil
		}
	}
	return "", fmt.Errorf("unsupported platform: %s/%s", goos, goarch)
}

// UpdateInfo holds metadata about an official Antigravity language_server release.
type UpdateInfo struct {
	Platform        string `json:"platform"`
	LatestVersion   string `json:"latest_version"`
	RawBuildVersion string `json:"raw_build_version"`
	DownloadURL     string `json:"download_url"`
	UpdateAvailable bool   `json:"update_available"`
}

// ResolveLatest queries the official download page and discovers the latest build for the current platform.
func ResolveLatest(pageURL string, goos, goarch string) (*UpdateInfo, error) {
	if pageURL == "" {
		pageURL = DefaultDownloadPage
	}

	slug, err := PlatformSlug(goos, goarch)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET", pageURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", "antigravity-remote-updater")
	req.Header.Set("Accept-Encoding", "gzip, deflate")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch download page %s: %w", pageURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download page returned HTTP %d", resp.StatusCode)
	}

	var reader io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gz, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("decompress gzip: %w", err)
		}
		defer gz.Close()
		reader = gz
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read download page body: %w", err)
	}

	matches := hubURLRe.FindAllString(string(body), -1)
	for _, m := range matches {
		if strings.HasSuffix(m, slug) {
			verMatch := hubVersionRe.FindStringSubmatch(m)
			version := "unknown"
			if len(verMatch) > 1 {
				version = verMatch[1]
			}

			return &UpdateInfo{
				Platform:      fmt.Sprintf("%s/%s", goos, goarch),
				LatestVersion: version,
				DownloadURL:   m,
			}, nil
		}
	}

	return nil, fmt.Errorf("no official download artifact found for %s (%s)", slug, goos+"/"+goarch)
}

// CheckUpdate compares the installed version with the latest official version.
func CheckUpdate(currentVersion string) (*UpdateInfo, error) {
	info, err := ResolveLatest(DefaultDownloadPage, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return nil, err
	}
	if currentVersion != "" && info.LatestVersion != "unknown" {
		info.UpdateAvailable = currentVersion != info.LatestVersion
	}
	return info, nil
}

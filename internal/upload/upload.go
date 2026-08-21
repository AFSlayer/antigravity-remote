package upload

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	// DefaultTTL is how long uploaded files are kept before automatic cleanup.
	DefaultTTL = 7 * 24 * time.Hour
	// UploadAPIPath is the endpoint for streaming file uploads.
	UploadAPIPath = "/__agy/api/upload"
)

// Options configures an Uploader.
type Options struct {
	WorkspaceRoot string
	TTL           time.Duration
}

// Uploader handles streaming file uploads to the workspace and periodic cleanup.
type Uploader struct {
	workspaceRoot string
	ttl           time.Duration
}

// New creates a new Uploader.
func New(opts Options) *Uploader {
	ttl := opts.TTL
	if ttl <= 0 {
		ttl = DefaultTTL
	}
	return &Uploader{
		workspaceRoot: opts.WorkspaceRoot,
		ttl:           ttl,
	}
}

// Register mounts upload endpoints on mux.
func (u *Uploader) Register(mux *http.ServeMux) {
	mux.HandleFunc(UploadAPIPath, u.handleUpload)
}

// StartCleaner runs a periodic background task to clean up files older than TTL.
func (u *Uploader) StartCleaner(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = time.Hour
	}

	go func() {
		// Run once on startup
		u.cleanupOldFiles()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				u.cleanupOldFiles()
			}
		}
	}()
}

// Response returned after successful upload.
type Response struct {
	Success      bool   `json:"success"`
	FileName     string `json:"fileName"`
	RelativePath string `json:"relativePath"`
	FullPath     string `json:"fullPath"`
	Size         int64  `json:"size"`
	Error        string `json:"error,omitempty"`
}

func (u *Uploader) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, Response{Error: "Method not allowed"})
		return
	}

	mr, err := r.MultipartReader()
	if err != nil {
		writeJSON(w, http.StatusBadRequest, Response{Error: fmt.Sprintf("invalid multipart request: %v", err)})
		return
	}

	var conversationID string
	var projectPath string
	var savedResponse *Response

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			writeJSON(w, http.StatusBadRequest, Response{Error: fmt.Sprintf("read part error: %v", err)})
			return
		}

		formName := part.FormName()
		if formName == "conversationId" {
			data, _ := io.ReadAll(io.LimitReader(part, 1024))
			conversationID = sanitizeID(string(data))
			continue
		}
		if formName == "projectPath" {
			data, _ := io.ReadAll(io.LimitReader(part, 4096))
			projectPath = strings.TrimSpace(string(data))
			continue
		}

		if formName == "file" || part.FileName() != "" {
			fileName := sanitizeFileName(part.FileName())
			if fileName == "" {
				fileName = fmt.Sprintf("upload_%d.bin", time.Now().Unix())
			}

			if conversationID == "" {
				conversationID = "shared"
			}

			targetDir, relativePrefix, err := u.resolveTargetDir(projectPath, conversationID)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, Response{Error: err.Error()})
				return
			}

			if err := os.MkdirAll(targetDir, 0755); err != nil {
				writeJSON(w, http.StatusInternalServerError, Response{Error: fmt.Sprintf("cannot create upload directory: %v", err)})
				return
			}

			destPath := filepath.Join(targetDir, fileName)
			outFile, err := os.Create(destPath)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, Response{Error: fmt.Sprintf("cannot create file: %v", err)})
				return
			}

			written, err := io.Copy(outFile, part)
			_ = outFile.Close()
			if err != nil {
				_ = os.Remove(destPath)
				writeJSON(w, http.StatusInternalServerError, Response{Error: fmt.Sprintf("write file error: %v", err)})
				return
			}

			relPath := filepath.Join(relativePrefix, fileName)
			savedResponse = &Response{
				Success:      true,
				FileName:     fileName,
				RelativePath: relPath,
				FullPath:     destPath,
				Size:         written,
			}
		}
	}

	if savedResponse == nil {
		writeJSON(w, http.StatusBadRequest, Response{Error: "no file was provided in the upload request"})
		return
	}

	writeJSON(w, http.StatusOK, *savedResponse)
}

func (u *Uploader) resolveTargetDir(projectPath, conversationID string) (targetDir string, relativePrefix string, err error) {
	wsRoot := u.workspaceRoot
	home, _ := os.UserHomeDir()
	if wsRoot == "" {
		wsRoot = filepath.Join(home, "antigravity", "workspace")
	}
	wsRoot = filepath.Clean(wsRoot)

	if projectPath != "" {
		cleanProj := filepath.Clean(projectPath)
		if !filepath.IsAbs(cleanProj) {
			cleanProj = filepath.Join(wsRoot, cleanProj)
		}
		// Security boundary check: ensure cleanProj does not escape wsRoot
		if rel, err := filepath.Rel(wsRoot, cleanProj); err == nil && !strings.HasPrefix(rel, "..") && rel != ".." {
			targetDir = filepath.Join(cleanProj, "uploads", conversationID)
			relativePrefix = filepath.Join("uploads", conversationID)
			return targetDir, relativePrefix, nil
		}
	}

	// Outside of project: store in workspace/uploads/{conversationID}
	targetDir = filepath.Join(wsRoot, "uploads", conversationID)

	// Calculate relative path from user's home directory so the agent can find it immediately from ~
	if home != "" && strings.HasPrefix(targetDir, home) {
		if rel, err := filepath.Rel(home, targetDir); err == nil {
			relativePrefix = rel
			return targetDir, relativePrefix, nil
		}
	}

	relativePrefix = filepath.Join("antigravity", "workspace", "uploads", conversationID)
	return targetDir, relativePrefix, nil
}

func (u *Uploader) cleanupOldFiles() {
	wsRoot := u.workspaceRoot
	if wsRoot == "" {
		home, _ := os.UserHomeDir()
		wsRoot = filepath.Join(home, "antigravity", "workspace")
	}

	if _, err := os.Stat(wsRoot); os.IsNotExist(err) {
		return
	}

	cutoff := time.Now().Add(-u.ttl)

	// Walk workspace and clean files inside uploads/ directories
	_ = filepath.Walk(wsRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			if strings.Contains(path, "/uploads/") || strings.Contains(path, "/.temporary_uploads/") {
				if info.ModTime().Before(cutoff) {
					if err := os.Remove(path); err == nil {
						log.Printf("[upload cleaner] deleted expired file: %s", path)
					}
				}
			}
		}
		return nil
	})

	// Remove empty directories in uploads
	uploadsRoot := filepath.Join(wsRoot, "uploads")
	removeEmptyDirs(uploadsRoot)
	tempRoot := filepath.Join(wsRoot, ".temporary_uploads")
	removeEmptyDirs(tempRoot)
}

func removeEmptyDirs(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if entry.IsDir() {
			sub := filepath.Join(dir, entry.Name())
			removeEmptyDirs(sub)
			_ = os.Remove(sub)
		}
	}
}

func sanitizeID(raw string) string {
	raw = strings.TrimSpace(raw)
	var b strings.Builder
	for _, r := range raw {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			b.WriteRune(r)
		}
	}
	res := b.String()
	if len(res) > 64 {
		res = res[:64]
	}
	return res
}

func sanitizeFileName(name string) string {
	name = filepath.Base(name)
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	if name == "" || name == "." || name == ".." {
		return fmt.Sprintf("upload_%d.bin", time.Now().Unix())
	}
	return name
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

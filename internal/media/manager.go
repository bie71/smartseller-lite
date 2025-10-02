package media

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"

	_ "golang.org/x/image/webp"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

const (
	productsDirName = "products"
	logosDirName    = "logos"
	masterMaxSize   = 1600
	thumbMaxSize    = 256
	logoMaxSize     = 512
)

// Asset represents a stored media file pair (master + optional thumbnail).
type Asset struct {
	Path           string
	ThumbPath      string
	Hash           string
	Width          int
	Height         int
	SizeBytes      int64
	ThumbWidth     int
	ThumbHeight    int
	ThumbSizeBytes int64
}

// Manager coordinates media persistence on disk and URL generation.
type Manager struct {
	baseDir  string
	mediaDir string
}

// ResolveBaseDir determines the SmartSellerLite data directory for the host OS.
func ResolveBaseDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}

	switch runtime.GOOS {
	case "windows":
		if appData := strings.TrimSpace(os.Getenv("APPDATA")); appData != "" {
			return filepath.Join(appData, "SmartSellerLite"), nil
		}
		return filepath.Join(home, "AppData", "Roaming", "SmartSellerLite"), nil
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "SmartSellerLite"), nil
	default:
		return filepath.Join(home, ".local", "share", "SmartSellerLite"), nil
	}
}

// NewManager constructs a Manager rooted at the provided base directory.
func NewManager(baseDir string) (*Manager, error) {
	if strings.TrimSpace(baseDir) == "" {
		return nil, errors.New("base directory is required")
	}
	mediaDir := filepath.Join(baseDir, "media")
	if err := os.MkdirAll(mediaDir, 0o755); err != nil {
		return nil, fmt.Errorf("create media dir: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(mediaDir, productsDirName), 0o755); err != nil {
		return nil, fmt.Errorf("prepare products dir: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(mediaDir, logosDirName), 0o755); err != nil {
		return nil, fmt.Errorf("prepare logos dir: %w", err)
	}
	return &Manager{baseDir: baseDir, mediaDir: mediaDir}, nil
}

// BaseDir returns the root data directory for the application.
func (m *Manager) BaseDir() string {
	return m.baseDir
}

// MediaDir returns the directory that houses all media assets.
func (m *Manager) MediaDir() string {
	return m.mediaDir
}

// PublicURL converts a relative media path to a HTTP-accessible URL.
func (m *Manager) PublicURL(rel string) string {
	clean := strings.TrimSpace(rel)
	if clean == "" {
		return ""
	}
	clean = filepath.ToSlash(clean)
	clean = strings.TrimPrefix(clean, "./")
	clean = strings.TrimPrefix(clean, "/")
	return "/media/" + clean
}

// SaveProductImage persists a product image, generating master + thumbnail variants.
func (m *Manager) SaveProductImage(ctx context.Context, payload string) (*Asset, error) {
	return m.saveImage(ctx, payload, productsDirName, masterMaxSize, thumbMaxSize, true)
}

// SaveLogo persists a logo image with a single optimised master variant.
func (m *Manager) SaveLogo(ctx context.Context, payload string) (*Asset, error) {
	return m.saveImage(ctx, payload, logosDirName, logoMaxSize, 0, false)
}

// Remove deletes the provided relative media paths if present.
func (m *Manager) Remove(relPaths ...string) error {
	for _, rel := range relPaths {
		clean := strings.TrimSpace(rel)
		if clean == "" {
			continue
		}
		target, err := m.safeJoin(clean)
		if err != nil {
			return err
		}
		if err := os.Remove(target); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("remove media %s: %w", clean, err)
		}
	}
	return nil
}

// Read returns the raw bytes of a stored media asset.
func (m *Manager) Read(rel string) ([]byte, error) {
	target, err := m.safeJoin(rel)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(target)
	if err != nil {
		return nil, fmt.Errorf("read media %s: %w", rel, err)
	}
	return data, nil
}

// ArchiveToZip writes the current media directory into the provided zip writer.
func (m *Manager) ArchiveToZip(z *zip.Writer) error {
	base := m.MediaDir()
	if _, err := os.Stat(base); errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return fmt.Errorf("stat media dir: %w", err)
	}

	return filepath.WalkDir(base, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(base, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(filepath.Join("media", rel))
		writer, err := z.Create(rel)
		if err != nil {
			return err
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(writer, file)
		closeErr := file.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
		return nil
	})
}

// RestoreFromZip extracts media files from the provided zip reader.
func (m *Manager) RestoreFromZip(reader *zip.Reader) error {
	for _, f := range reader.File {
		if f.FileInfo().IsDir() {
			continue
		}
		if !strings.HasPrefix(f.Name, "media/") {
			continue
		}
		rel := strings.TrimPrefix(f.Name, "media/")
		if rel == "" || rel == "." {
			continue
		}
		target, err := m.safeJoin(rel)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return fmt.Errorf("prepare media dir: %w", err)
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return err
		}
		if err := os.WriteFile(target, data, 0o644); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) saveImage(ctx context.Context, payload, bucket string, maxSize, thumbSize int, includeThumb bool) (*Asset, error) {
	if ctx != nil && ctx.Err() != nil {
		return nil, ctx.Err()
	}
	data, err := decodeBase64Image(payload)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, errors.New("empty image payload")
	}

	img, err := imaging.Decode(bytes.NewReader(data), imaging.AutoOrientation(true))
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}

	master := imaging.Fit(img, maxSize, maxSize, imaging.Lanczos)
	masterBytes, masterExt, err := encodeImageData(master, 90)
	if err != nil {
		return nil, fmt.Errorf("encode master image: %w", err)
	}

	asset := &Asset{
		Width:     master.Bounds().Dx(),
		Height:    master.Bounds().Dy(),
		SizeBytes: int64(len(masterBytes)),
		Hash:      fmt.Sprintf("%x", sha256.Sum256(masterBytes)),
	}

	now := time.Now().UTC()
	dir := filepath.Join(m.MediaDir(), bucket, fmt.Sprintf("%04d", now.Year()), fmt.Sprintf("%02d", int(now.Month())))
	if bucket == logosDirName {
		dir = filepath.Join(m.MediaDir(), bucket)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create media dir: %w", err)
	}
	id := uuid.New().String()
	masterName := fmt.Sprintf("%s%s", id, masterExt)
	masterPath := filepath.Join(bucket, masterName)
	if bucket != logosDirName {
		masterPath = filepath.Join(bucket, fmt.Sprintf("%04d", now.Year()), fmt.Sprintf("%02d", int(now.Month())), masterName)
	}

	if err := os.WriteFile(filepath.Join(m.MediaDir(), masterPath), masterBytes, 0o644); err != nil {
		return nil, fmt.Errorf("write master image: %w", err)
	}
	asset.Path = filepath.ToSlash(masterPath)

	if includeThumb && thumbSize > 0 {
		thumb := imaging.Fit(img, thumbSize, thumbSize, imaging.Lanczos)
		thumbBytes, thumbExt, err := encodeImageData(thumb, 85)
		if err != nil {
			return nil, fmt.Errorf("encode thumbnail: %w", err)
		}
		thumbName := fmt.Sprintf("%s.thumb%s", id, thumbExt)
		thumbPath := filepath.Join(bucket, fmt.Sprintf("%04d", now.Year()), fmt.Sprintf("%02d", int(now.Month())), thumbName)
		if err := os.WriteFile(filepath.Join(m.MediaDir(), thumbPath), thumbBytes, 0o644); err != nil {
			return nil, fmt.Errorf("write thumbnail: %w", err)
		}
		asset.ThumbPath = filepath.ToSlash(thumbPath)
		asset.ThumbWidth = thumb.Bounds().Dx()
		asset.ThumbHeight = thumb.Bounds().Dy()
		asset.ThumbSizeBytes = int64(len(thumbBytes))
	}

	return asset, nil
}

func (m *Manager) safeJoin(rel string) (string, error) {
	clean := filepath.Clean(rel)
	clean = strings.TrimPrefix(clean, "./")
	clean = strings.TrimPrefix(clean, string(os.PathSeparator))
	if clean == "." || clean == "" {
		return "", fmt.Errorf("invalid media path: %s", rel)
	}
	target := filepath.Join(m.MediaDir(), clean)
	abs, err := filepath.Abs(target)
	if err != nil {
		return "", err
	}
	mediaAbs, err := filepath.Abs(m.MediaDir())
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(abs, mediaAbs) {
		return "", fmt.Errorf("media path escapes base: %s", rel)
	}
	return abs, nil
}

func decodeBase64Image(payload string) ([]byte, error) {
	data := strings.TrimSpace(payload)
	if data == "" {
		return nil, errors.New("image payload required")
	}
	if strings.HasPrefix(data, "data:") {
		if idx := strings.Index(data, ";base64,"); idx != -1 {
			data = data[idx+8:]
		} else if idx := strings.Index(data, ","); idx != -1 {
			data = data[idx+1:]
		}
	}
	data = strings.TrimSpace(data)
	data = strings.ReplaceAll(data, "\n", "")
	data = strings.ReplaceAll(data, "\r", "")
	if mod := len(data) % 4; mod != 0 {
		data += strings.Repeat("=", 4-mod)
	}
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, fmt.Errorf("decode base64 image: %w", err)
	}
	if len(decoded) == 0 {
		return nil, errors.New("decoded image empty")
	}
	if http.DetectContentType(decoded[:min(len(decoded), 512)]) == "application/octet-stream" {
		return decoded, nil
	}
	return decoded, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// CreateArchive produces a zip archive containing the provided SQL dump alongside media files.
func (m *Manager) CreateArchive(sqlName string, sqlData []byte, includeMedia bool) ([]byte, error) {
	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)
	sqlFile := sqlName
	if strings.TrimSpace(sqlFile) == "" {
		sqlFile = "dump.sql"
	}
	file, err := writer.Create(sqlFile)
	if err != nil {
		writer.Close()
		return nil, err
	}
	if _, err := file.Write(sqlData); err != nil {
		writer.Close()
		return nil, err
	}
	if includeMedia {
		if err := m.ArchiveToZip(writer); err != nil {
			writer.Close()
			return nil, err
		}
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// RestoreArchive extracts dump bytes and media contents from the provided archive payload.
func (m *Manager) RestoreArchive(payload []byte) ([]byte, error) {
	reader, err := zip.NewReader(bytes.NewReader(payload), int64(len(payload)))
	if err != nil {
		return nil, err
	}
	var dump []byte
	for _, f := range reader.File {
		if f.FileInfo().IsDir() {
			continue
		}
		name := strings.ToLower(filepath.Base(f.Name))
		if name == "dump.sql" || strings.HasSuffix(name, ".sql") {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			dump, err = io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return nil, err
			}
			continue
		}
	}
	if len(dump) == 0 {
		return nil, errors.New("archive missing SQL dump")
	}
	if err := m.RestoreFromZip(reader); err != nil {
		return nil, err
	}
	return dump, nil
}

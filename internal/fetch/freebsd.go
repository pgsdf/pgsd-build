package fetch

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pgsdf/pgsdbuild/internal/util"
)

// Fetcher fetches FreeBSD distribution archives from mirrors.
type Fetcher struct {
	version string
	arch    string
	mirror  string
	destDir string
	logger  *util.Logger
}

// NewFetcher creates a new FreeBSD archive fetcher.
func NewFetcher(version, arch, mirror, destDir string, logger *util.Logger) *Fetcher {
	// Default to official FreeBSD mirror if not specified
	if mirror == "" {
		mirror = "https://download.freebsd.org"
	}

	return &Fetcher{
		version: version,
		arch:    arch,
		mirror:  mirror,
		destDir: destDir,
		logger:  logger,
	}
}

// FetchArchives downloads base.txz and kernel.txz if they don't exist locally.
// Returns the paths to the archives.
func (f *Fetcher) FetchArchives() (basePath, kernelPath string, err error) {
	f.logger.Info("Checking for FreeBSD %s (%s) distribution archives...", f.version, f.arch)

	// Ensure destination directory exists
	if err := util.EnsureDir(f.destDir); err != nil {
		return "", "", fmt.Errorf("failed to create destination directory: %w", err)
	}

	basePath = filepath.Join(f.destDir, "base.txz")
	kernelPath = filepath.Join(f.destDir, "kernel.txz")

	// Check if archives already exist
	baseExists := util.FileExists(basePath)
	kernelExists := util.FileExists(kernelPath)

	if baseExists && kernelExists {
		f.logger.Info("Archives already cached in %s", f.destDir)

		// Verify they're not empty or corrupt
		if err := f.verifyArchive(basePath); err != nil {
			f.logger.Warn("Cached base.txz appears corrupt, will re-download: %v", err)
			baseExists = false
		}
		if err := f.verifyArchive(kernelPath); err != nil {
			f.logger.Warn("Cached kernel.txz appears corrupt, will re-download: %v", err)
			kernelExists = false
		}

		if baseExists && kernelExists {
			return basePath, kernelPath, nil
		}
	}

	// Build mirror URL
	// Format: https://download.freebsd.org/releases/amd64/14.2-RELEASE/base.txz
	baseURL := fmt.Sprintf("%s/releases/%s/%s", f.mirror, f.arch, f.version)

	// Download missing archives
	if !baseExists {
		f.logger.Info("Downloading base.txz from %s...", baseURL)
		if err := f.downloadFile(baseURL+"/base.txz", basePath); err != nil {
			return "", "", fmt.Errorf("failed to download base.txz: %w", err)
		}
		f.logger.Info("Downloaded base.txz successfully")
	}

	if !kernelExists {
		f.logger.Info("Downloading kernel.txz from %s...", baseURL)
		if err := f.downloadFile(baseURL+"/kernel.txz", kernelPath); err != nil {
			return "", "", fmt.Errorf("failed to download kernel.txz: %w", err)
		}
		f.logger.Info("Downloaded kernel.txz successfully")
	}

	// Optional: Download and verify checksums
	if err := f.downloadAndVerifyChecksums(baseURL, basePath, kernelPath); err != nil {
		f.logger.Warn("Checksum verification failed or unavailable: %v", err)
		f.logger.Warn("Continuing anyway - archives may be corrupt")
	}

	return basePath, kernelPath, nil
}

// downloadFile downloads a file from URL to destination with progress reporting.
func (f *Fetcher) downloadFile(url, destPath string) error {
	// Create temporary file
	tmpPath := destPath + ".tmp"
	defer os.Remove(tmpPath) // Clean up on error

	// Create output file
	out, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	// Send HTTP GET request
	client := &http.Client{
		Timeout: 30 * time.Minute, // Large files may take time
	}

	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: HTTP %d %s", resp.StatusCode, resp.Status)
	}

	// Get file size for progress reporting
	totalSize := resp.ContentLength

	// Copy with progress reporting
	var written int64
	buf := make([]byte, 32*1024) // 32KB buffer
	lastReport := time.Now()

	for {
		nr, err := resp.Body.Read(buf)
		if nr > 0 {
			nw, err := out.Write(buf[0:nr])
			if err != nil {
				return fmt.Errorf("failed to write file: %w", err)
			}
			if nw != nr {
				return fmt.Errorf("short write")
			}
			written += int64(nw)

			// Report progress every 2 seconds
			if time.Since(lastReport) > 2*time.Second {
				if totalSize > 0 {
					percent := float64(written) / float64(totalSize) * 100
					f.logger.Info("  Downloaded: %.1f%% (%.1f MB / %.1f MB)",
						percent,
						float64(written)/(1024*1024),
						float64(totalSize)/(1024*1024))
				} else {
					f.logger.Info("  Downloaded: %.1f MB", float64(written)/(1024*1024))
				}
				lastReport = time.Now()
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to read: %w", err)
		}
	}

	// Close and sync
	if err := out.Sync(); err != nil {
		return fmt.Errorf("failed to sync file: %w", err)
	}
	if err := out.Close(); err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}

	// Move temporary file to final destination
	if err := os.Rename(tmpPath, destPath); err != nil {
		return fmt.Errorf("failed to rename file: %w", err)
	}

	f.logger.Info("  Download complete: %.1f MB", float64(written)/(1024*1024))

	return nil
}

// verifyArchive performs basic verification on a .txz archive.
func (f *Fetcher) verifyArchive(path string) error {
	// Check file exists and is not empty
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if info.Size() < 1024 {
		return fmt.Errorf("file too small (possibly corrupt): %d bytes", info.Size())
	}

	// Basic check: verify it starts with xz magic bytes
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	magic := make([]byte, 6)
	if _, err := file.Read(magic); err != nil {
		return err
	}

	// xz magic: 0xFD 0x37 0x7A 0x58 0x5A 0x00
	if magic[0] != 0xFD || magic[1] != 0x37 || magic[2] != 0x7A ||
		magic[3] != 0x58 || magic[4] != 0x5A || magic[5] != 0x00 {
		return fmt.Errorf("not a valid xz archive")
	}

	return nil
}

// downloadAndVerifyChecksums downloads MANIFEST and verifies checksums.
func (f *Fetcher) downloadAndVerifyChecksums(baseURL, basePath, kernelPath string) error {
	// Download MANIFEST file
	manifestURL := baseURL + "/MANIFEST"

	f.logger.Debug("Downloading MANIFEST for checksum verification...")

	// Download MANIFEST
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(manifestURL)
	if err != nil {
		return fmt.Errorf("failed to download MANIFEST: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("MANIFEST not found: HTTP %d", resp.StatusCode)
	}

	// Parse MANIFEST
	checksums := make(map[string]string)
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		// Format: base.txz SHA256 (base.txz) = <hash>
		// Or: SHA256 (base.txz) = <hash>
		if strings.Contains(line, "SHA256") && strings.Contains(line, " = ") {
			parts := strings.Split(line, " = ")
			if len(parts) == 2 {
				// Extract filename from SHA256 (filename)
				filenamePart := strings.TrimSpace(parts[0])
				if start := strings.Index(filenamePart, "("); start >= 0 {
					if end := strings.Index(filenamePart, ")"); end > start {
						filename := filenamePart[start+1 : end]
						hash := strings.TrimSpace(parts[1])
						checksums[filename] = hash
					}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to parse MANIFEST: %w", err)
	}

	// Verify checksums
	f.logger.Debug("Verifying checksums...")

	if hash, ok := checksums["base.txz"]; ok {
		if err := f.verifyChecksum(basePath, hash); err != nil {
			return fmt.Errorf("base.txz checksum mismatch: %w", err)
		}
		f.logger.Info("✓ base.txz checksum verified")
	} else {
		f.logger.Warn("No checksum found for base.txz in MANIFEST")
	}

	if hash, ok := checksums["kernel.txz"]; ok {
		if err := f.verifyChecksum(kernelPath, hash); err != nil {
			return fmt.Errorf("kernel.txz checksum mismatch: %w", err)
		}
		f.logger.Info("✓ kernel.txz checksum verified")
	} else {
		f.logger.Warn("No checksum found for kernel.txz in MANIFEST")
	}

	return nil
}

// verifyChecksum verifies a file's SHA256 checksum.
func (f *Fetcher) verifyChecksum(path, expectedHash string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return err
	}

	actualHash := hex.EncodeToString(hash.Sum(nil))

	if actualHash != expectedHash {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedHash, actualHash)
	}

	return nil
}

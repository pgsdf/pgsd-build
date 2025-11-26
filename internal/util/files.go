package util

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// CopyFile copies a single file from src to dst with the specified mode.
func CopyFile(src, dst string, mode os.FileMode) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read source file %s: %w", src, err)
	}
	if err := os.WriteFile(dst, data, mode); err != nil {
		return fmt.Errorf("failed to write destination file %s: %w", dst, err)
	}
	return nil
}

// CopyDir recursively copies a directory from src to dst.
// It handles symlinks by copying them as symlinks and gracefully skips broken symlinks.
func CopyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return fmt.Errorf("failed to calculate relative path: %w", err)
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		// Handle symlinks specially
		if info.Mode()&os.ModeSymlink != 0 {
			// Read the symlink target
			linkTarget, err := os.Readlink(path)
			if err != nil {
				// If we can't read the symlink, skip it
				return nil
			}

			// Create the symlink at the destination
			// Remove existing file/link if present
			os.Remove(dstPath)
			if err := os.Symlink(linkTarget, dstPath); err != nil {
				// If symlink creation fails, just skip it
				// This can happen with broken symlinks
				return nil
			}
			return nil
		}

		// Copy regular file
		return CopyFile(path, dstPath, info.Mode())
	})
}

// CopyOverlay recursively copies overlay files from src to dst.
// This is similar to CopyDir but specifically for overlay operations.
func CopyOverlay(src, dst string) error {
	return CopyDir(src, dst)
}

// EnsureDir creates a directory if it doesn't exist.
func EnsureDir(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", path, err)
	}
	return nil
}

// FileExists checks if a file exists and is not a directory.
func FileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// DirExists checks if a directory exists.
func DirExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// CleanupDir removes a directory and all its contents.
// It tries to handle root-owned files by:
// 1. Attempting normal removal
// 2. If that fails, trying to change permissions recursively
// 3. If that fails, trying to use sudo
// 4. Returning an error if all methods fail
func CleanupDir(path string) error {
	// Try normal removal first
	err := os.RemoveAll(path)
	if err == nil {
		return nil
	}

	// If removal failed, try changing permissions first
	_ = filepath.Walk(path, func(file string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Ignore errors during walk
		}
		// Try to make everything writable and executable
		_ = os.Chmod(file, 0777)
		return nil
	})

	// Try removal again after chmod
	err = os.RemoveAll(path)
	if err == nil {
		return nil
	}

	// If still failing, try using sudo (if available)
	return CleanupDirWithSudo(path)
}

// CleanupDirWithSudo attempts to remove a directory using sudo.
// Returns nil if successful, error if sudo is not available or fails.
func CleanupDirWithSudo(path string) error {
	// Check if sudo is available
	sudoPath, err := exec.LookPath("sudo")
	if err != nil {
		// Sudo not available, return original error
		return fmt.Errorf("cannot remove directory %s (permission denied, sudo not available)", path)
	}

	// Try to remove with sudo
	cmd := exec.Command(sudoPath, "rm", "-rf", path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("sudo rm failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// CreateSparseFile creates a sparse file of the specified size.
func CreateSparseFile(path string, size int64) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", path, err)
	}
	defer f.Close()

	if err := f.Truncate(size); err != nil {
		return fmt.Errorf("failed to truncate file to %d bytes: %w", size, err)
	}

	return nil
}

// WriteStringToFile writes a string to a file, creating it if it doesn't exist.
func WriteStringToFile(path string, content string, perm os.FileMode) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", path, err)
	}
	defer f.Close()

	if _, err := io.WriteString(f, content); err != nil {
		return fmt.Errorf("failed to write to file %s: %w", path, err)
	}

	if err := f.Chmod(perm); err != nil {
		return fmt.Errorf("failed to set permissions on %s: %w", path, err)
	}

	return nil
}

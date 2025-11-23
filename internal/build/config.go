package build

import (
	"os"
	"path/filepath"
)

// Config holds the build system configuration.
type Config struct {
	// Directory paths
	ImagesDir    string
	VariantsDir  string
	ArtifactsDir string
	WorkDir      string
	ISODir       string
	OverlaysDir  string

	// Build options
	Verbose    bool
	KeepWork   bool
	DiskSizeGB int

	// Runtime paths
	RootDir string
}

// NewDefaultConfig creates a Config with default values.
func NewDefaultConfig() *Config {
	return &Config{
		ImagesDir:    "images",
		VariantsDir:  "variants",
		ArtifactsDir: "artifacts",
		WorkDir:      "work",
		ISODir:       "iso",
		OverlaysDir:  "overlays",
		Verbose:      false,
		KeepWork:     false,
		DiskSizeGB:   10,
		RootDir:      ".",
	}
}

// ResolveDir resolves a directory path relative to RootDir.
func (c *Config) ResolveDir(dir string) string {
	if filepath.IsAbs(dir) {
		return dir
	}
	return filepath.Join(c.RootDir, dir)
}

// GetImagesDir returns the absolute path to the images directory.
func (c *Config) GetImagesDir() string {
	return c.ResolveDir(c.ImagesDir)
}

// GetVariantsDir returns the absolute path to the variants directory.
func (c *Config) GetVariantsDir() string {
	return c.ResolveDir(c.VariantsDir)
}

// GetArtifactsDir returns the absolute path to the artifacts directory.
func (c *Config) GetArtifactsDir() string {
	return c.ResolveDir(c.ArtifactsDir)
}

// GetWorkDir returns the absolute path to the work directory.
func (c *Config) GetWorkDir() string {
	return c.ResolveDir(c.WorkDir)
}

// GetISODir returns the absolute path to the ISO directory.
func (c *Config) GetISODir() string {
	return c.ResolveDir(c.ISODir)
}

// GetOverlaysDir returns the absolute path to the overlays directory.
func (c *Config) GetOverlaysDir() string {
	return c.ResolveDir(c.OverlaysDir)
}

// LoadFromEnv loads configuration from environment variables.
func (c *Config) LoadFromEnv() {
	if v := os.Getenv("PGSD_IMAGES_DIR"); v != "" {
		c.ImagesDir = v
	}
	if v := os.Getenv("PGSD_VARIANTS_DIR"); v != "" {
		c.VariantsDir = v
	}
	if v := os.Getenv("PGSD_ARTIFACTS_DIR"); v != "" {
		c.ArtifactsDir = v
	}
	if v := os.Getenv("PGSD_WORK_DIR"); v != "" {
		c.WorkDir = v
	}
	if v := os.Getenv("PGSD_ISO_DIR"); v != "" {
		c.ISODir = v
	}
	if v := os.Getenv("PGSD_OVERLAYS_DIR"); v != "" {
		c.OverlaysDir = v
	}
	if v := os.Getenv("PGSD_VERBOSE"); v == "1" || v == "true" {
		c.Verbose = true
	}
	if v := os.Getenv("PGSD_KEEP_WORK"); v == "1" || v == "true" {
		c.KeepWork = true
	}
}

// Validate checks that the configuration is valid.
func (c *Config) Validate() error {
	// All validation is currently optional since directories are created as needed
	return nil
}

package iso

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pgsdf/pgsdbuild/internal/build"
	"github.com/pgsdf/pgsdbuild/internal/config"
	"github.com/pgsdf/pgsdbuild/internal/util"
)

// Builder builds bootable ISO images.
type Builder struct {
	config *build.Config
	logger *util.Logger
}

// NewBuilder creates a new ISO Builder.
func NewBuilder(cfg *build.Config, logger *util.Logger) *Builder {
	return &Builder{
		config: cfg,
		logger: logger,
	}
}

// Build implements the boot environment ISO build pipeline.
func (b *Builder) Build(cfg config.VariantConfig) error {
	b.logger.Info("Starting bootenv ISO build for %s", cfg.ID)

	// Create working directories
	workPath := filepath.Join(b.config.GetWorkDir(), "iso", cfg.ID)
	if err := util.EnsureDir(workPath); err != nil {
		return err
	}

	if !b.config.KeepWork {
		defer util.CleanupDir(workPath)
	}

	outputPath := filepath.Join(b.config.GetISODir(), cfg.ID+".iso")
	if err := util.EnsureDir(b.config.GetISODir()); err != nil {
		return err
	}

	// Step 1: Build root filesystem for ISO
	isoRoot := filepath.Join(workPath, "root")
	if err := util.EnsureDir(isoRoot); err != nil {
		return err
	}

	b.logger.Debug("Building root filesystem...")
	if err := b.buildISOFilesystem(cfg, isoRoot); err != nil {
		return fmt.Errorf("failed to build ISO filesystem: %w", err)
	}

	// Step 2: Install packages
	b.logger.Debug("Installing packages...")
	if err := b.installISOPackages(cfg, isoRoot); err != nil {
		return fmt.Errorf("failed to install packages: %w", err)
	}

	// Step 3: Apply overlays
	b.logger.Debug("Applying overlays...")
	if err := b.applyISOOverlays(cfg, isoRoot); err != nil {
		return fmt.Errorf("failed to apply overlays: %w", err)
	}

	// Step 4: Copy system images
	if cfg.ImagesDir != "" {
		b.logger.Debug("Copying system images...")
		if err := b.copySystemImages(cfg, isoRoot); err != nil {
			return fmt.Errorf("failed to copy system images: %w", err)
		}
	}

	// Step 5: Register Arcan target
	b.logger.Debug("Registering Arcan installer target...")
	if err := b.registerArcanTarget(isoRoot); err != nil {
		return fmt.Errorf("failed to register Arcan target: %w", err)
	}

	// Step 6: Assemble ISO image
	b.logger.Debug("Assembling ISO image...")
	if err := b.assembleISO(cfg, isoRoot, outputPath); err != nil {
		return fmt.Errorf("failed to assemble ISO: %w", err)
	}

	b.logger.Info("ISO build complete! Output: %s", outputPath)
	return nil
}

// buildISOFilesystem creates the base directory structure for the ISO.
func (b *Builder) buildISOFilesystem(cfg config.VariantConfig, isoRoot string) error {
	// Create standard FreeBSD directory structure
	dirs := []string{
		"bin", "boot", "dev", "etc", "lib", "libexec",
		"mnt", "proc", "rescue", "root", "sbin", "tmp",
		"usr/bin", "usr/lib", "usr/local/bin", "usr/local/etc",
		"usr/local/lib", "usr/local/share", "usr/sbin",
		"usr/share", "var/log", "var/run", "var/tmp",
	}

	for _, dir := range dirs {
		path := filepath.Join(isoRoot, dir)
		if err := util.EnsureDir(path); err != nil {
			return err
		}
	}

	b.logger.Debug("Created base filesystem structure")
	return nil
}

// installISOPackages installs packages into the ISO root.
func (b *Builder) installISOPackages(cfg config.VariantConfig, isoRoot string) error {
	// On FreeBSD:
	// pkg -r isoRoot install -y <packages>

	// For prototype, create a marker file showing what packages would be installed
	pkgList := filepath.Join(isoRoot, "installed-packages.txt")
	content := fmt.Sprintf("# Bootenv ISO: %s\n# Package sets installed:\n", cfg.ID)
	for _, pkgSet := range cfg.PkgLists {
		content += fmt.Sprintf("# - %s\n", pkgSet)
	}

	if err := util.WriteStringToFile(pkgList, content, 0644); err != nil {
		return err
	}

	b.logger.Debug("Installed package lists: %v", cfg.PkgLists)
	return nil
}

// applyISOOverlays applies filesystem overlays to the ISO root.
func (b *Builder) applyISOOverlays(cfg config.VariantConfig, isoRoot string) error {
	overlaysDir := b.config.GetOverlaysDir()

	for _, overlay := range cfg.Overlays {
		overlayPath := filepath.Join(overlaysDir, overlay)

		// Check if overlay exists
		if !util.DirExists(overlayPath) {
			return fmt.Errorf("overlay %s not found at %s", overlay, overlayPath)
		}

		// Copy overlay contents to isoRoot
		if err := util.CopyOverlay(overlayPath, isoRoot); err != nil {
			return fmt.Errorf("failed to copy overlay %s: %w", overlay, err)
		}

		b.logger.Debug("Applied overlay: %s", overlay)
	}
	return nil
}

// copySystemImages copies built system images into the ISO.
func (b *Builder) copySystemImages(cfg config.VariantConfig, isoRoot string) error {
	imagesDestDir := filepath.Join(isoRoot, cfg.ImagesDir[1:]) // Remove leading /
	if err := util.EnsureDir(imagesDestDir); err != nil {
		return err
	}

	// Look for artifacts in the artifacts directory
	artifactsDir := b.config.GetArtifactsDir()
	entries, err := os.ReadDir(artifactsDir)
	if err != nil {
		// If no artifacts directory exists, that's okay
		if os.IsNotExist(err) {
			b.logger.Warn("No system images found in %s", artifactsDir)
			return nil
		}
		return fmt.Errorf("failed to read artifacts directory: %w", err)
	}

	imageCount := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		imageName := entry.Name()
		imageArtifactDir := filepath.Join(artifactsDir, imageName)
		imageDestDir := filepath.Join(imagesDestDir, imageName)

		// Copy the entire artifact directory
		if err := util.CopyDir(imageArtifactDir, imageDestDir); err != nil {
			return fmt.Errorf("failed to copy image %s: %w", imageName, err)
		}

		b.logger.Debug("Copied system image: %s", imageName)
		imageCount++
	}

	b.logger.Info("Copied %d system image(s) to %s", imageCount, cfg.ImagesDir)
	return nil
}

// registerArcanTarget creates the Arcan target registration metadata.
func (b *Builder) registerArcanTarget(isoRoot string) error {
	// On a real system, this would use arcan_db to register the target
	// For prototype, we'll create a marker file

	arcanDir := filepath.Join(isoRoot, "usr/local/share/arcan")
	if err := util.EnsureDir(arcanDir); err != nil {
		return err
	}

	targetFile := filepath.Join(arcanDir, "pgsd-installer-target.txt")
	content := fmt.Sprintf("# PGSD Installer Arcan Target\n"+
		"# Registered at: %s\n\n"+
		"target_name: pgsd-installer\n"+
		"target_type: BINARY\n"+
		"target_path: /usr/local/bin/pgsd-inst\n"+
		"config: default\n",
		time.Now().Format(time.RFC3339))

	if err := util.WriteStringToFile(targetFile, content, 0644); err != nil {
		return err
	}

	b.logger.Debug("Registered pgsd-installer as Arcan target")
	return nil
}

// assembleISO creates the final ISO image.
func (b *Builder) assembleISO(cfg config.VariantConfig, isoRoot, outputPath string) error {
	// On FreeBSD, we would use:
	// makefs -t cd9660 -o rockridge -o label=PGSD_BOOT outputPath isoRoot

	// For prototype, create a marker ISO file
	content := fmt.Sprintf("# PGSD Bootenv ISO (prototype)\n"+
		"# Variant: %s (%s)\n"+
		"# Created: %s\n"+
		"# Package lists: %v\n"+
		"# Overlays: %v\n"+
		"# Images dir: %s\n",
		cfg.ID, cfg.Name,
		time.Now().Format(time.RFC3339),
		cfg.PkgLists,
		cfg.Overlays,
		cfg.ImagesDir)

	if err := util.WriteStringToFile(outputPath, content, 0644); err != nil {
		return err
	}

	b.logger.Debug("Created ISO image: %s", outputPath)
	return nil
}

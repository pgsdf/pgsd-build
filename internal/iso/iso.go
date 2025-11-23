package iso

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

	// Step 2.5: Install FreeBSD boot infrastructure
	b.logger.Debug("Installing boot infrastructure...")
	if err := b.installBootInfrastructure(isoRoot); err != nil {
		return fmt.Errorf("failed to install boot infrastructure: %w", err)
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
	// Install FreeBSD packages into the ISO root filesystem
	// Uses pkg with -r flag to install to alternate root

	if len(cfg.PkgLists) == 0 {
		b.logger.Warn("No package lists specified for ISO variant")
		return nil
	}

	// TODO: Expand package lists from cfg.PkgLists to actual package names
	// For now, we'll need to implement package list resolution
	// This would involve reading package list files and expanding them

	// Example for when package resolution is implemented:
	// args := []string{"-r", isoRoot, "install", "-y"}
	// for _, pkg := range resolvedPackages {
	//     args = append(args, pkg)
	// }
	// if err := b.runCommand("pkg", args...); err != nil {
	//     return fmt.Errorf("package installation failed: %w", err)
	// }

	b.logger.Info("Package installation: %d package lists configured", len(cfg.PkgLists))
	b.logger.Debug("Package lists: %v", cfg.PkgLists)

	// Create a marker file documenting what should be installed
	pkgList := filepath.Join(isoRoot, "PACKAGES.txt")
	content := fmt.Sprintf("Bootenv ISO: %s\nPackage sets configured:\n", cfg.ID)
	for _, pkgSet := range cfg.PkgLists {
		content += fmt.Sprintf("  - %s\n", pkgSet)
	}
	content += "\nNote: Package installation requires FreeBSD pkg and package list resolution.\n"

	if err := util.WriteStringToFile(pkgList, content, 0644); err != nil {
		return err
	}

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

// installBootInfrastructure installs FreeBSD boot files needed for bootable ISO
func (b *Builder) installBootInfrastructure(isoRoot string) error {
	// Copy essential boot files from the running FreeBSD system
	// This makes the ISO bootable in both BIOS and UEFI modes

	bootDir := filepath.Join(isoRoot, "boot")

	// Required boot files for BIOS boot
	biosBootFiles := []string{
		"/boot/cdboot",      // CD/DVD boot loader
		"/boot/loader",      // Boot loader
		"/boot/loader.rc",   // Loader configuration
		"/boot/defaults/loader.conf", // Default loader settings
	}

	b.logger.Debug("Copying BIOS boot files...")
	for _, srcPath := range biosBootFiles {
		if _, err := os.Stat(srcPath); err != nil {
			if os.IsNotExist(err) {
				b.logger.Warn("Boot file not found (skipping): %s", srcPath)
				continue
			}
			return fmt.Errorf("failed to access boot file %s: %w", srcPath, err)
		}

		// Determine destination path
		relPath := strings.TrimPrefix(srcPath, "/boot/")
		dstPath := filepath.Join(bootDir, relPath)

		// Ensure destination directory exists
		dstDir := filepath.Dir(dstPath)
		if err := util.EnsureDir(dstDir); err != nil {
			return err
		}

		// Copy the file
		if err := util.CopyFile(srcPath, dstPath, 0644); err != nil {
			return fmt.Errorf("failed to copy %s: %w", srcPath, err)
		}
	}

	// Copy kernel (required for boot)
	b.logger.Debug("Copying FreeBSD kernel...")
	kernelSrc := "/boot/kernel/kernel"
	if _, err := os.Stat(kernelSrc); err == nil {
		kernelDir := filepath.Join(bootDir, "kernel")
		if err := util.EnsureDir(kernelDir); err != nil {
			return err
		}
		kernelDst := filepath.Join(kernelDir, "kernel")
		if err := util.CopyFile(kernelSrc, kernelDst, 0755); err != nil {
			return fmt.Errorf("failed to copy kernel: %w", err)
		}
	} else {
		b.logger.Warn("Kernel not found at %s - ISO may not boot", kernelSrc)
	}

	// Copy essential kernel modules
	b.logger.Debug("Copying kernel modules...")
	modules := []string{
		"zfs.ko",       // ZFS filesystem
		"geom_label.ko", // GEOM labels
		"ahci.ko",      // AHCI disk controller
	}

	for _, module := range modules {
		srcPath := filepath.Join("/boot/kernel", module)
		if _, err := os.Stat(srcPath); err == nil {
			dstPath := filepath.Join(bootDir, "kernel", module)
			if err := util.CopyFile(srcPath, dstPath, 0644); err != nil {
				b.logger.Warn("Failed to copy module %s: %v", module, err)
			}
		}
	}

	// Set up EFI boot directory for UEFI support
	b.logger.Debug("Setting up EFI boot...")
	efiDir := filepath.Join(isoRoot, "EFI", "BOOT")
	if err := util.EnsureDir(efiDir); err != nil {
		return err
	}

	// Copy EFI boot loader if available
	efiBootSrc := "/boot/boot1.efi"
	if _, err := os.Stat(efiBootSrc); err == nil {
		efiBootDst := filepath.Join(efiDir, "BOOTX64.EFI")
		if err := util.CopyFile(efiBootSrc, efiBootDst, 0644); err != nil {
			b.logger.Warn("Failed to copy EFI bootloader: %v", err)
		}
	} else {
		b.logger.Warn("EFI bootloader not found - UEFI boot may not work")
	}

	b.logger.Info("Boot infrastructure installed")
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
	// Create bootable ISO using makefs
	// This requires FreeBSD with makefs utility

	// Determine ISO label (max 32 characters, alphanumeric only for cd9660)
	// makefs cd9660 requires "d-characters" (digits/letters) only in labels
	label := cfg.ID
	label = strings.ReplaceAll(label, "-", "")
	label = strings.ReplaceAll(label, "_", "")
	if len(label) > 32 {
		label = label[:32]
	}
	// Ensure label is uppercase for consistency
	label = strings.ToUpper(label)

	b.logger.Debug("Creating ISO filesystem...")
	b.logger.Debug("ISO label: %s", label)

	// Check if cdboot file exists before configuring boot
	cdbootPath := filepath.Join(isoRoot, "boot/cdboot")
	hasCdboot := false
	if stat, err := os.Stat(cdbootPath); err == nil && !stat.IsDir() {
		hasCdboot = true
		b.logger.Info("BIOS boot file found: boot/cdboot (size: %d bytes)", stat.Size())
	} else {
		b.logger.Warn("BIOS boot file not found at %s - ISO will NOT be bootable in BIOS mode", cdbootPath)
		b.logger.Warn("To create bootable ISOs, this must run on FreeBSD with boot files installed")
	}

	// Detect which ISO creation tool is available
	isoTool, isoToolPath := b.detectISOTool()
	if isoTool == "" {
		b.logger.Warn("No ISO creation tool found (tried: makefs, xorriso, genisoimage, mkisofs)")
		b.logger.Warn("Creating tar archive instead - convert to ISO on a system with ISO tools")

		// Fallback: create a tar.gz archive of the ISO contents
		tarPath := strings.TrimSuffix(outputPath, ".iso") + ".tar.gz"
		if err := b.createTarArchive(tarPath, isoRoot); err != nil {
			return fmt.Errorf("tar archive creation failed: %w", err)
		}

		b.logger.Info("Created tar archive: %s", tarPath)
		b.logger.Info("To convert to ISO, use: genisoimage -r -V %s -o %s -graft-points <extracted-contents>", label, outputPath)

		return nil
	}

	b.logger.Info("Using ISO creation tool: %s (%s)", isoTool, isoToolPath)

	// Convert paths to absolute to avoid working directory issues
	absIsoRoot, err := filepath.Abs(isoRoot)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for ISO root: %w", err)
	}

	absOutputPath, err := filepath.Abs(outputPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for output: %w", err)
	}

	b.logger.Debug("ISO root (absolute): %s", absIsoRoot)
	b.logger.Debug("Output path (absolute): %s", absOutputPath)

	// Create ISO using the detected tool
	if err := b.createISOWithTool(isoTool, isoToolPath, absOutputPath, absIsoRoot, label, hasCdboot); err != nil {
		return fmt.Errorf("%s failed: %w", isoTool, err)
	}

	b.logger.Debug("Created ISO image: %s", outputPath)

	// Verify the ISO was created
	if info, err := os.Stat(outputPath); err != nil {
		return fmt.Errorf("ISO verification failed: %w", err)
	} else {
		b.logger.Info("ISO size: %.2f MB", float64(info.Size())/(1024*1024))
	}

	return nil
}

// detectISOTool detects which ISO creation tool is available on the system
// Returns the tool name and full path to the executable
func (b *Builder) detectISOTool() (string, string) {
	// Try common tools with explicit paths for FreeBSD/Linux systems
	toolPaths := map[string][]string{
		"makefs":      {"/usr/sbin/makefs", "/sbin/makefs"},
		"xorriso":     {"/usr/bin/xorriso"},
		"genisoimage": {"/usr/bin/genisoimage"},
		"mkisofs":     {"/usr/bin/mkisofs"},
	}

	// Check explicit paths first (for FreeBSD where tools may not be in PATH)
	for tool, paths := range toolPaths {
		for _, path := range paths {
			if _, err := os.Stat(path); err == nil {
				b.logger.Debug("Found ISO tool at %s", path)
				return tool, path
			}
		}
	}

	// Fall back to PATH search
	tools := []string{"makefs", "xorriso", "genisoimage", "mkisofs"}
	for _, tool := range tools {
		if path, err := exec.LookPath(tool); err == nil {
			return tool, path
		}
	}

	return "", ""
}

// createISOWithTool creates an ISO using the specified tool
func (b *Builder) createISOWithTool(tool, toolPath, outputPath, isoRoot, label string, hasCdboot bool) error {
	switch tool {
	case "makefs":
		return b.createISOWithMakefs(toolPath, outputPath, isoRoot, label, hasCdboot)
	case "xorriso":
		return b.createISOWithXorriso(toolPath, outputPath, isoRoot, label, hasCdboot)
	case "genisoimage", "mkisofs":
		return b.createISOWithGenisoimage(toolPath, outputPath, isoRoot, label, hasCdboot)
	default:
		return fmt.Errorf("unsupported ISO tool: %s", tool)
	}
}

// createISOWithMakefs creates an ISO using FreeBSD's makefs utility
func (b *Builder) createISOWithMakefs(toolPath, outputPath, isoRoot, label string, hasCdboot bool) error {
	// Build makefs command arguments
	// -t cd9660: ISO 9660 filesystem
	// -o rockridge (R): Rock Ridge extensions (long filenames, permissions)
	// -o L=<label>: Volume label (must be d-characters: alphanumeric only)
	// -o B=<bootimage>: Boot image for BIOS boot
	// -o no-emul-boot: No emulation boot mode
	// -o no-trailing-padding: Omit padding for smaller file
	args := []string{
		"-t", "cd9660",
		"-o", "rockridge",
		"-o", fmt.Sprintf("L=%s", label),
	}

	// Add boot image configuration if cdboot exists
	if hasCdboot {
		args = append(args,
			"-o", "B=i386;boot/cdboot", // Boot image specification
			"-o", "no-emul-boot",        // No emulation mode
		)
		b.logger.Info("Configured for BIOS boot with boot/cdboot")
	} else {
		b.logger.Info("Creating non-bootable ISO (no boot files available)")
	}

	// Add final options and paths
	args = append(args,
		"-o", "no-trailing-padding",
		outputPath,
		isoRoot,
	)

	return b.runCommand(toolPath, args...)
}

// createISOWithXorriso creates an ISO using xorriso (modern Linux ISO tool)
func (b *Builder) createISOWithXorriso(toolPath, outputPath, isoRoot, label string, hasCdboot bool) error {
	// xorriso command arguments
	// -as mkisofs: Compatibility mode
	// -r: Rock Ridge extensions
	// -V <label>: Volume label
	// -o <output>: Output file
	args := []string{
		"-as", "mkisofs",
		"-r",                       // Rock Ridge
		"-V", label,                // Volume label
		"-o", outputPath,           // Output file
		"-graft-points",            // Enable graft points for file mapping
	}

	if hasCdboot {
		// Note: xorriso boot configuration is complex and may need adjustment
		// for FreeBSD boot files. This creates a basic bootable ISO.
		args = append(args,
			"-b", "boot/cdboot",     // Boot image
			"-no-emul-boot",         // No emulation
			"-boot-load-size", "4",  // Load size
			"-boot-info-table",      // Create boot info table
		)
		b.logger.Info("Configured for BIOS boot with boot/cdboot")
	} else {
		b.logger.Info("Creating non-bootable ISO (no boot files available)")
	}

	args = append(args, isoRoot)

	return b.runCommand(toolPath, args...)
}

// createISOWithGenisoimage creates an ISO using genisoimage or mkisofs (legacy Linux tools)
func (b *Builder) createISOWithGenisoimage(toolPath, outputPath, isoRoot, label string, hasCdboot bool) error {
	// genisoimage/mkisofs command arguments
	// -r: Rock Ridge extensions
	// -V <label>: Volume label
	// -o <output>: Output file
	args := []string{
		"-r",                // Rock Ridge
		"-V", label,         // Volume label
		"-o", outputPath,    // Output file
	}

	if hasCdboot {
		args = append(args,
			"-b", "boot/cdboot",     // Boot image
			"-no-emul-boot",         // No emulation
			"-boot-load-size", "4",  // Load size
			"-boot-info-table",      // Create boot info table
		)
		b.logger.Info("Configured for BIOS boot with boot/cdboot")
	} else {
		b.logger.Info("Creating non-bootable ISO (no boot files available)")
	}

	args = append(args, isoRoot)

	return b.runCommand(toolPath, args...)
}

// createTarArchive creates a compressed tar archive of the ISO root directory
func (b *Builder) createTarArchive(tarPath, isoRoot string) error {
	b.logger.Debug("Creating tar archive: %s", tarPath)

	// Create the output file
	outFile, err := os.Create(tarPath)
	if err != nil {
		return fmt.Errorf("failed to create tar file: %w", err)
	}
	defer outFile.Close()

	// Create gzip writer
	gzWriter := gzip.NewWriter(outFile)
	defer gzWriter.Close()

	// Create tar writer
	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// Walk the ISO root directory and add all files
	return filepath.Walk(isoRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get the relative path for the tar header
		relPath, err := filepath.Rel(isoRoot, path)
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if relPath == "." {
			return nil
		}

		// Create tar header from file info
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return fmt.Errorf("failed to create tar header for %s: %w", path, err)
		}

		// Use the relative path as the name in the archive
		header.Name = relPath

		// Write the header
		if err := tarWriter.WriteHeader(header); err != nil {
			return fmt.Errorf("failed to write tar header for %s: %w", path, err)
		}

		// If it's a regular file, write its contents
		if info.Mode().IsRegular() {
			file, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("failed to open file %s: %w", path, err)
			}
			defer file.Close()

			if _, err := io.Copy(tarWriter, file); err != nil {
				return fmt.Errorf("failed to write file %s to tar: %w", path, err)
			}
		}

		return nil
	})
}

// runCommand executes a command and returns an error if it fails
func (b *Builder) runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)

	b.logger.Debug("Running: %s %v", name, args)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command failed: %w\nOutput: %s", err, string(output))
	}

	if len(output) > 0 {
		b.logger.Debug("Command output: %s", string(output))
	}

	return nil
}
